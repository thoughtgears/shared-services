package telemetry

type Otel struct {
	ServiceName string
	DomainName  string
	Endpoint    string
}

func NewTelemetry(serviceName, domainName, endpoint string) *Otel {
	return &Otel{
		ServiceName: serviceName,
		DomainName:  domainName,
		Endpoint:    endpoint,
	}
}
