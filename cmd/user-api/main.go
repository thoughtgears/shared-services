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
	"github.com/thoughtgears/shared-services/pkg/telemetry"
)

var cfg config.Config

func init() {
	envconfig.MustProcess("", &cfg)
	zerolog.SetGlobalLevel(zerolog.InfoLevel)
	zerolog.LevelFieldName = "severity"
}

func main() {
	ctx := context.Background()

	otel := telemetry.NewTelemetry(cfg.ServiceName, cfg.DomainName, cfg.OTELEndpoint)
	cleanup := otel.InitTracer(ctx)
	defer func() {
		if err := cleanup(ctx); err != nil {
			log.Fatal().Msgf("Failed to cleanup OpenTelemetry: %v", err)
		}
	}()

	shutdown := otel.InitCounter(ctx)
	defer func() {
		if err := shutdown(ctx); err != nil {
			log.Fatal().Msgf("Failed to shutdown OpenTelemetry: %v", err)
		}
	}()

	client, err := firestore.NewClient(ctx, cfg.ProjectID)
	if err != nil {
		log.Fatal().Msgf("Failed to create Firestore client: %v", err)
	}

	datastore := db.NewFirestoreRepository[models.User](client, cfg.FirstoreCollection)
	userService := services.NewUserService(datastore)
	userHandler := handlers.NewHandler(userService)

	r := router.NewRouter(cfg.ServiceName, cfg.Local, &cfg.Port)

	userHandler.RegisterRoutes(r.Engine)

	log.Fatal().Err(r.Run()).Msg("Failed to run server")
}
