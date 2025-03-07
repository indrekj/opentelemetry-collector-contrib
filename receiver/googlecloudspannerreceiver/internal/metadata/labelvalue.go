// Copyright  The OpenTelemetry Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package metadata // import "github.com/open-telemetry/opentelemetry-collector-contrib/receiver/googlecloudspannerreceiver/internal/metadata"

import (
	"fmt"
	"sort"
	"strings"

	"go.opentelemetry.io/collector/model/pdata"
)

type newLabelValueFunction func(m LabelValueMetadata, value interface{}) LabelValue

type LabelValueMetadata interface {
	ValueMetadata
	ValueType() ValueType
	NewLabelValue(value interface{}) LabelValue
}

type LabelValue interface {
	Metadata() LabelValueMetadata
	Value() interface{}
	SetValueTo(attributes pdata.AttributeMap)
}

type queryLabelValueMetadata struct {
	name              string
	columnName        string
	valueType         ValueType
	newLabelValueFunc newLabelValueFunction
	valueHolderFunc   valueHolderFunction
}

func (m queryLabelValueMetadata) ValueHolder() interface{} {
	return m.valueHolderFunc()
}

func (m queryLabelValueMetadata) NewLabelValue(value interface{}) LabelValue {
	return m.newLabelValueFunc(m, value)
}

func (m queryLabelValueMetadata) ValueType() ValueType {
	return m.valueType
}

type stringLabelValue struct {
	metadata LabelValueMetadata
	value    string
}

type int64LabelValue struct {
	metadata LabelValueMetadata
	value    int64
}

type boolLabelValue struct {
	metadata LabelValueMetadata
	value    bool
}

type stringSliceLabelValue struct {
	metadata LabelValueMetadata
	value    string
}

type byteSliceLabelValue struct {
	metadata LabelValueMetadata
	value    string
}

func (m queryLabelValueMetadata) Name() string {
	return m.name
}

func (m queryLabelValueMetadata) ColumnName() string {
	return m.columnName
}

func (v stringLabelValue) Metadata() LabelValueMetadata {
	return v.metadata
}

func (v stringLabelValue) Value() interface{} {
	return v.value
}

func (v stringLabelValue) SetValueTo(attributes pdata.AttributeMap) {
	attributes.InsertString(v.metadata.Name(), v.value)
}

func newStringLabelValue(metadata LabelValueMetadata, valueHolder interface{}) LabelValue {
	return stringLabelValue{
		metadata: metadata,
		value:    *valueHolder.(*string),
	}
}

func (v int64LabelValue) Metadata() LabelValueMetadata {
	return v.metadata
}

func (v int64LabelValue) Value() interface{} {
	return v.value
}

func (v int64LabelValue) SetValueTo(attributes pdata.AttributeMap) {
	attributes.InsertInt(v.metadata.Name(), v.value)
}

func newInt64LabelValue(metadata LabelValueMetadata, valueHolder interface{}) LabelValue {
	return int64LabelValue{
		metadata: metadata,
		value:    *valueHolder.(*int64),
	}
}

func (v boolLabelValue) Metadata() LabelValueMetadata {
	return v.metadata
}

func (v boolLabelValue) Value() interface{} {
	return v.value
}

func (v boolLabelValue) SetValueTo(attributes pdata.AttributeMap) {
	attributes.InsertBool(v.metadata.Name(), v.value)
}

func newBoolLabelValue(metadata LabelValueMetadata, valueHolder interface{}) LabelValue {
	return boolLabelValue{
		metadata: metadata,
		value:    *valueHolder.(*bool),
	}
}

func (v stringSliceLabelValue) Metadata() LabelValueMetadata {
	return v.metadata
}

func (v stringSliceLabelValue) Value() interface{} {
	return v.value
}

func (v stringSliceLabelValue) SetValueTo(attributes pdata.AttributeMap) {
	attributes.InsertString(v.metadata.Name(), v.value)
}

func newStringSliceLabelValue(metadata LabelValueMetadata, valueHolder interface{}) LabelValue {
	value := *valueHolder.(*[]string)

	sort.Strings(value)

	sortedAndConstructedValue := strings.Join(value, ",")

	return stringSliceLabelValue{
		metadata: metadata,
		value:    sortedAndConstructedValue,
	}
}

func (v byteSliceLabelValue) Metadata() LabelValueMetadata {
	return v.metadata
}

func (v byteSliceLabelValue) Value() interface{} {
	return v.value
}

func (v byteSliceLabelValue) SetValueTo(attributes pdata.AttributeMap) {
	attributes.InsertString(v.metadata.Name(), v.value)
}

func newByteSliceLabelValue(metadata LabelValueMetadata, valueHolder interface{}) LabelValue {
	return byteSliceLabelValue{
		metadata: metadata,
		value:    string(*valueHolder.(*[]byte)),
	}
}

func NewLabelValueMetadata(name string, columnName string, valueType ValueType) (LabelValueMetadata, error) {
	var newLabelValueFunc newLabelValueFunction
	var valueHolderFunc valueHolderFunction

	switch valueType {
	case StringValueType:
		newLabelValueFunc = newStringLabelValue
		valueHolderFunc = func() interface{} {
			var valueHolder string
			return &valueHolder
		}
	case IntValueType:
		newLabelValueFunc = newInt64LabelValue
		valueHolderFunc = func() interface{} {
			var valueHolder int64
			return &valueHolder
		}
	case BoolValueType:
		newLabelValueFunc = newBoolLabelValue
		valueHolderFunc = func() interface{} {
			var valueHolder bool
			return &valueHolder
		}
	case StringSliceValueType:
		newLabelValueFunc = newStringSliceLabelValue
		valueHolderFunc = func() interface{} {
			var valueHolder []string
			return &valueHolder
		}
	case ByteSliceValueType:
		newLabelValueFunc = newByteSliceLabelValue
		valueHolderFunc = func() interface{} {
			var valueHolder []byte
			return &valueHolder
		}
	default:
		return nil, fmt.Errorf("invalid value type received for label %q", name)
	}

	return queryLabelValueMetadata{
		name:              name,
		columnName:        columnName,
		valueType:         valueType,
		newLabelValueFunc: newLabelValueFunc,
		valueHolderFunc:   valueHolderFunc,
	}, nil
}
