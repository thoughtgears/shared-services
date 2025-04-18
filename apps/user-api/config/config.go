package config

type Config struct {
	ProjectID          string `envconfig:"GCP_PROJECT_ID" required:"true"`
	Region             string `envconfig:"GCP_REGION" required:"true"`
	Local              bool   `envconfig:"LOCAL" default:"false"`
	Port               string `envconfig:"PORT" default:"8080"`
	FirstoreCollection string `envconfig:"FIRESTORE_COLLECTION" default:"users"`
	ServiceName        string `envconfig:"K_SERVICE" default:"user-api"`
	DomainName         string `envconfig:"DOMAIN_NAME" default:"thoughtgears.co.uk"`
}
