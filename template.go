package otchkiss

const defaultReportTemplate = `
[Setting]
* warm up time:   {{.WarmUpTime}}
* duration:       {{.Duration}}
* max concurrent: {{.MaxConcurrent}}
* max RPS:        {{.MaxRPS}}

[Request]
* total:      {{.TotalRequests}}
* succeeded:  {{.Succeeded}}
* failed:     {{.Failed}}
* error rate: {{.ErrorRate}} %
* RPS:        {{.RPS}}

[Latency]
* max: {{.MaxLatency}} ms
* min: {{.MinLatency}} ms
* avg: {{.AvgLatency}} ms
* med: {{.MedLatency}} ms
* 99th percentile: {{.Latency99p}} ms
* 90th percentile: {{.Latency90p}} ms
`
