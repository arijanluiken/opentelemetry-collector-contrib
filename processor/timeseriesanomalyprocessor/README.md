# Timeseries Anomaly Processor

Detect anomalies in numeric time series (metrics) using a simple rolling z-score model. It maintains a fixed-size window per series and flags a point as anomalous when its absolute z-score exceeds the configured threshold.

- Signal support: metrics (Gauge, Sum)
- Modes:
  - enrich: annotate points with anomaly.score and anomaly.flag
  - filter: drop non-anomalous points
  - both: keep only anomalies and annotate them

## Configuration

- window_size (int, default 50): number of recent points per series
- zscore_threshold (float, default 3.0): abs(z) >= threshold => anomaly
- mode (string, default "enrich"): one of enrich, filter, both
- score_attribute (string, default "anomaly.score")
- flag_attribute (string, default "anomaly.flag")
- max_series (int, default 10000): limit tracked series; 0 = unlimited
- gc_interval (duration, default 5m): periodic cleanup of idle series
- idle_ttl (duration, default 15m): prune series idle longer than this

## Example

```yaml
processors:
  timeseriesanomaly:
    window_size: 60
    zscore_threshold: 3.5
    mode: both
    score_attribute: anomaly.score
    flag_attribute: anomaly.flag

service:
  pipelines:
    metrics:
      receivers: [otlp]
      processors: [timeseriesanomaly]
      exporters: [debug]
```

## Notes

- Series key is derived from metric name, resource attributes, and datapoint attributes.
- Only numeric number datapoints (int/double) are analyzed; other metric types pass through unchanged.
- This processor is stateful and memory usage grows with the number of active series.
