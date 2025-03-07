// Copyright The OpenTelemetry Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//       http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

//go:build linux
// +build linux

package diskscraper // import "github.com/open-telemetry/opentelemetry-collector-contrib/receiver/hostmetricsreceiver/internal/scraper/diskscraper"

import (
	"github.com/shirou/gopsutil/disk"
	"go.opentelemetry.io/collector/model/pdata"

	"github.com/open-telemetry/opentelemetry-collector-contrib/receiver/hostmetricsreceiver/internal/scraper/diskscraper/internal/metadata"
)

const systemSpecificMetricsLen = 2

func appendSystemSpecificMetrics(metrics pdata.MetricSlice, startTime, now pdata.Timestamp, ioCounters map[string]disk.IOCountersStat) {
	initializeDiskWeightedIOTimeMetric(metrics.AppendEmpty(), startTime, now, ioCounters)
	initializeDiskMergedMetric(metrics.AppendEmpty(), startTime, now, ioCounters)
}

func initializeDiskWeightedIOTimeMetric(metric pdata.Metric, startTime, now pdata.Timestamp, ioCounters map[string]disk.IOCountersStat) {
	metadata.Metrics.SystemDiskWeightedIoTime.Init(metric)

	ddps := metric.Sum().DataPoints()
	ddps.EnsureCapacity(len(ioCounters))

	for device, ioCounter := range ioCounters {
		initializeNumberDataPointAsDouble(ddps.AppendEmpty(), startTime, now, device, "", float64(ioCounter.WeightedIO)/1e3)
	}
}

func initializeDiskMergedMetric(metric pdata.Metric, startTime, now pdata.Timestamp, ioCounters map[string]disk.IOCountersStat) {
	metadata.Metrics.SystemDiskMerged.Init(metric)

	idps := metric.Sum().DataPoints()
	idps.EnsureCapacity(2 * len(ioCounters))

	for device, ioCounter := range ioCounters {
		initializeNumberDataPointAsInt(idps.AppendEmpty(), startTime, now, device, metadata.AttributeDirection.Read, int64(ioCounter.MergedReadCount))
		initializeNumberDataPointAsInt(idps.AppendEmpty(), startTime, now, device, metadata.AttributeDirection.Write, int64(ioCounter.MergedWriteCount))
	}
}
