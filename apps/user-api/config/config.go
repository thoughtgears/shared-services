package config

type Config struct {
	ProjectID string `envconfig:"GCP_PROJECT_ID" required:"true"`
	Region    string `envconfig:"GCP_REGION" required:"true"`
	Local     bool   `envconfig:"LOCAL" default:"false"`
	Port      string `envconfig:"PORT" default:"8080"`
}
