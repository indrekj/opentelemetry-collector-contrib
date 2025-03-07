// Copyright The OpenTelemetry Authors
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

package utils // import "github.com/open-telemetry/opentelemetry-collector-contrib/exporter/datadogexporter/internal/utils"

import (
	"crypto/tls"
	"fmt"
	"net"
	"net/http"
	"time"

	"go.opentelemetry.io/collector/component"
)

var (
	// JSONHeaders headers for JSON requests.
	JSONHeaders = map[string]string{
		"Content-Type":     "application/json",
		"Content-Encoding": "gzip",
	}
	// ProtobufHeaders headers for protobuf requests.
	ProtobufHeaders = map[string]string{
		"Content-Type":     "application/x-protobuf",
		"Content-Encoding": "identity",
	}
)

// NewHTTPClient returns a http.Client configured with the Agent options.
func NewHTTPClient(timeout time.Duration) *http.Client {
	return &http.Client{
		Timeout: timeout,
		Transport: &http.Transport{
			Proxy: http.ProxyFromEnvironment,
			DialContext: (&net.Dialer{
				// Disable RFC 6555 Fast Fallback ("Happy Eyeballs")
				FallbackDelay: -1 * time.Nanosecond,
			}).DialContext,
			MaxIdleConns: 100,
			// Not supported by intake
			ForceAttemptHTTP2: false,
			TLSClientConfig:   &tls.Config{InsecureSkipVerify: false},
		},
	}
}

// SetExtraHeaders appends a header map to HTTP headers.
func SetExtraHeaders(h http.Header, extras map[string]string) {
	for key, value := range extras {
		h.Set(key, value)
	}
}

func UserAgent(buildInfo component.BuildInfo) string {
	return fmt.Sprintf("%s/%s", buildInfo.Command, buildInfo.Version)
}

// SetDDHeaders sets the Datadog-specific headers
func SetDDHeaders(reqHeader http.Header, buildInfo component.BuildInfo, apiKey string) {
	reqHeader.Set("DD-Api-Key", apiKey)
	reqHeader.Set("User-Agent", UserAgent(buildInfo))
}

// DoWithRetries repeats a fallible action up to `maxRetries` times
// with exponential backoff
func DoWithRetries(maxRetries int, fn func() error) (i int, err error) {
	wait := 1 * time.Second
	for i = 0; i < maxRetries; i++ {
		err = fn()
		if err == nil {
			return
		}
		time.Sleep(wait)
		wait = 2 * wait
	}

	return
}
