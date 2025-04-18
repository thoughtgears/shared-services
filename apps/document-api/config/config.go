package config

type Config struct {
	ProjectID           string `envconfig:"GCP_PROJECT_ID" required:"true"`
	Region              string `envconfig:"GCP_REGION" required:"true"`
	Local               bool   `envconfig:"LOCAL" default:"false"`
	Port                string `envconfig:"PORT" default:"8080"`
	BucketName          string `envconfig:"GCP_BUCKET_NAME" required:"true"`
	FirestoreCollection string `envconfig:"FIRESTORE_COLLECTION" default:"documents"`
	ServiceName         string `envconfig:"K_SERVICE" default:"document-api"`
	OtelEndpoint        string `envconfig:"OTEL_EXPORTER_OTLP_TRACES_ENDPOINT" default:"http://localhost:4317"`
	DomainName          string `envconfig:"DOMAIN_NAME" default:"thoughtgears.co.uk"`
}
