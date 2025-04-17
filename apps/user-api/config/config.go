package config

type Config struct {
	ProjectID    string `envconfig:"GCP_PROJECT_ID" required:"true"`
	Region       string `envconfig:"GCP_REGION" required:"true"`
	Local        bool   `envconfig:"LOCAL" default:"false"`
	Port         string `envconfig:"PORT" default:"8080"`
	ServiceName  string `envconfig:"K_SERVICE" default:"user-api"`
	OtelEndpoint string `envconfig:"OTEL_EXPORTER_OTLP_ENDPOINT" default:"http://localhost:4317"`
	OtelInsecrue bool   `envconfig:"OTEL_INSECURE" default:"true"`
	DomainName   string `envconfig:"DOMAIN_NAME" default:"thoughtgears.co.uk"`
}
