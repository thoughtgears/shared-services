package main

import (
	"context"

	"cloud.google.com/go/firestore"
	"github.com/kelseyhightower/envconfig"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"github.com/thoughtgears/shared-services/apps/user-api/config"
	"github.com/thoughtgears/shared-services/apps/user-api/handlers"
	"github.com/thoughtgears/shared-services/apps/user-api/services"
	"github.com/thoughtgears/shared-services/pkg/db"
	"github.com/thoughtgears/shared-services/pkg/models"
	"github.com/thoughtgears/shared-services/pkg/router"
)

const collection = "users"

var cfg config.Config

func init() {
	envconfig.MustProcess("", &cfg)
	zerolog.SetGlobalLevel(zerolog.InfoLevel)
	zerolog.LevelFieldName = "severity"
}

func main() {
	ctx := context.Background()
	client, err := firestore.NewClient(ctx, cfg.ProjectID)
	if err != nil {
		log.Fatal().Msgf("Failed to create Firestore client: %v", err)
	}

	datastore := db.NewFirestoreRepository[models.User](client, collection)
	userService := services.NewUserService(datastore)
	userHandler := handlers.NewHandler(userService)

	r := router.NewRouter(cfg.Local, &cfg.Port)

	userHandler.RegisterRoutes(r.Engine)

	log.Fatal().Err(r.Run()).Msg("Failed to run server")
}
