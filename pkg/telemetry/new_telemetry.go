package telemetry

type Otel struct {
	ServiceName  string
	DomainName   string
	CollectorURL string
}

func NewTelemetry(serviceName, domainName string) *Otel {
	return &Otel{
		ServiceName: serviceName,
		DomainName:  domainName,
	}
}
