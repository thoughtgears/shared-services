package telemetry

type Otel struct {
	ServiceName  string
	DomainName   string
	CollectorURL string
	Insecure     bool
}

func NewTelemetry(serviceName, domainName, collectorURL string, insecure bool) *Otel {
	return &Otel{
		ServiceName:  serviceName,
		DomainName:   domainName,
		CollectorURL: collectorURL,
		Insecure:     insecure,
	}
}
