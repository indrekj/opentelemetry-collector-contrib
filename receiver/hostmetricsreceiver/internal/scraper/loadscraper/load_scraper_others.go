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

//go:build !windows
// +build !windows

package loadscraper // import "github.com/open-telemetry/opentelemetry-collector-contrib/receiver/hostmetricsreceiver/internal/scraper/loadscraper"

import (
	"context"

	"github.com/shirou/gopsutil/load"
	"go.uber.org/zap"
)

// unix based systems sample & compute load averages in the kernel, so nothing to do here
func startSampling(_ context.Context, _ *zap.Logger) error {
	return nil
}

func stopSampling(_ context.Context) error {
	return nil
}

func getSampledLoadAverages() (*load.AvgStat, error) {
	return load.Avg()
}
