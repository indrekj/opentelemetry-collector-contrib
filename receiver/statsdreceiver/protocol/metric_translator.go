// Copyright 2020, OpenTelemetry Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package protocol // import "github.com/open-telemetry/opentelemetry-collector-contrib/receiver/statsdreceiver/protocol"

import (
	"sort"
	"time"

	"go.opentelemetry.io/collector/model/pdata"
	"gonum.org/v1/gonum/stat"
)

var (
	statsDDefaultPercentiles = []float64{0, 10, 50, 90, 95, 100}
)

func buildCounterMetric(parsedMetric statsDMetric, isMonotonicCounter bool, timeNow, lastIntervalTime time.Time) pdata.InstrumentationLibraryMetrics {
	ilm := pdata.NewInstrumentationLibraryMetrics()
	nm := ilm.Metrics().AppendEmpty()
	nm.SetName(parsedMetric.description.name)
	if parsedMetric.unit != "" {
		nm.SetUnit(parsedMetric.unit)
	}
	nm.SetDataType(pdata.MetricDataTypeSum)

	nm.Sum().SetAggregationTemporality(pdata.MetricAggregationTemporalityDelta)
	nm.Sum().SetIsMonotonic(isMonotonicCounter)

	dp := nm.Sum().DataPoints().AppendEmpty()
	dp.SetIntVal(parsedMetric.counterValue())
	dp.SetStartTimestamp(pdata.NewTimestampFromTime(lastIntervalTime))
	dp.SetTimestamp(pdata.NewTimestampFromTime(timeNow))
	for i, key := range parsedMetric.labelKeys {
		dp.Attributes().InsertString(key, parsedMetric.labelValues[i])
	}

	return ilm
}

func buildGaugeMetric(parsedMetric statsDMetric, timeNow time.Time) pdata.InstrumentationLibraryMetrics {
	ilm := pdata.NewInstrumentationLibraryMetrics()
	nm := ilm.Metrics().AppendEmpty()
	nm.SetName(parsedMetric.description.name)
	if parsedMetric.unit != "" {
		nm.SetUnit(parsedMetric.unit)
	}
	nm.SetDataType(pdata.MetricDataTypeGauge)
	dp := nm.Gauge().DataPoints().AppendEmpty()
	dp.SetDoubleVal(parsedMetric.gaugeValue())
	dp.SetTimestamp(pdata.NewTimestampFromTime(timeNow))
	for i, key := range parsedMetric.labelKeys {
		dp.Attributes().InsertString(key, parsedMetric.labelValues[i])
	}

	return ilm
}

func buildSummaryMetric(summary summaryMetric, startTime, timeNow time.Time, percentiles []float64, ilm pdata.InstrumentationLibraryMetrics) {
	nm := ilm.Metrics().AppendEmpty()
	nm.SetName(summary.name)
	nm.SetDataType(pdata.MetricDataTypeSummary)

	dp := nm.Summary().DataPoints().AppendEmpty()

	count := float64(0)
	sum := float64(0)
	for i := range summary.points {
		c := summary.weights[i]
		count += c
		sum += summary.points[i] * c
	}

	// Note: count is rounded here, see note in counterValue().
	dp.SetCount(uint64(count))
	dp.SetSum(sum)

	dp.SetStartTimestamp(pdata.NewTimestampFromTime(startTime))
	dp.SetTimestamp(pdata.NewTimestampFromTime(timeNow))
	for i, key := range summary.labelKeys {
		dp.Attributes().InsertString(key, summary.labelValues[i])
	}

	sort.Sort(dualSorter{summary.points, summary.weights})

	for _, pct := range percentiles {
		eachQuantile := dp.QuantileValues().AppendEmpty()
		eachQuantile.SetQuantile(pct / 100)
		eachQuantile.SetValue(stat.Quantile(pct/100, stat.Empirical, summary.points, summary.weights))
	}
}

func (s statsDMetric) counterValue() int64 {
	x := s.asFloat
	// Note statds counters are always represented as integers.
	// There is no statsd specification that says what should or
	// shouldn't be done here.  Rounding may occur for sample
	// rates that are not integer reciprocals.  Recommendation:
	// use integer reciprocal sampling rates.
	if 0 < s.sampleRate && s.sampleRate < 1 {
		x = x / s.sampleRate
	}
	return int64(x)
}

func (s statsDMetric) gaugeValue() float64 {
	// sampleRate does not have effect for gauge points.
	return s.asFloat
}

func (s statsDMetric) summaryValue() summaryRaw {
	count := 1.0
	if 0 < s.sampleRate && s.sampleRate < 1 {
		count /= s.sampleRate
	}
	return summaryRaw{
		value: s.asFloat,
		count: count,
	}
}

type dualSorter struct {
	values, weights []float64
}

func (d dualSorter) Len() int {
	return len(d.values)
}

func (d dualSorter) Swap(i, j int) {
	d.values[i], d.values[j] = d.values[j], d.values[i]
	d.weights[i], d.weights[j] = d.weights[j], d.weights[i]
}

func (d dualSorter) Less(i, j int) bool {
	return d.values[i] < d.values[j]
}
