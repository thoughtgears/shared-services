package config

type Config struct {
	ProjectID    string `envconfig:"GCP_PROJECT_ID" required:"true"`
	Region       string `envconfig:"GCP_REGION" required:"true"`
	Local        bool   `envconfig:"LOCAL" default:"false"`
	Port         string `envconfig:"PORT" default:"8080"`
	BucketName   string `envconfig:"GCP_BUCKET_NAME" required:"true"`
	ServiceName  string `envconfig:"K_SERVICE" default:"portal-api"`
	DomainName   string `envconfig:"DOMAIN_NAME" default:"thoughtgears.co.uk"`
	OTELEndpoint string `envconfig:"OTEL_ENDPOINT" default:"localhost:4317"`
}
