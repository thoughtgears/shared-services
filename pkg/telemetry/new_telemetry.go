package telemetry

type Otel struct {
	ServiceName  string
	DomainName   string
	CollectorURL string
}

func NewTelemetry(serviceName, domainName, collectorURL string) *Otel {
	return &Otel{
		ServiceName:  serviceName,
		DomainName:   domainName,
		CollectorURL: collectorURL,
	}
}
