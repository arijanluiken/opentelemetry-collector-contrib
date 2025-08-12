// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

//go:generate mdatagen metadata.yaml

// package timeseriesanomalyprocessor implements a processor which detects
// anomalies in numeric time series (metrics) using simple online statistics.
package timeseriesanomalyprocessor // import "github.com/open-telemetry/opentelemetry-collector-contrib/processor/timeseriesanomalyprocessor"
