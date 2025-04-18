package main

import (
	"context"

	"cloud.google.com/go/firestore"
	"cloud.google.com/go/storage"
	"github.com/kelseyhightower/envconfig"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"github.com/thoughtgears/shared-services/internal/config"
	"github.com/thoughtgears/shared-services/internal/db"
	"github.com/thoughtgears/shared-services/internal/gcs"
	"github.com/thoughtgears/shared-services/internal/handlers"
	"github.com/thoughtgears/shared-services/internal/models"
	"github.com/thoughtgears/shared-services/internal/router"
	"github.com/thoughtgears/shared-services/internal/services"
	"github.com/thoughtgears/shared-services/internal/telemetry"
)

var cfg config.Config

const (
	userCollection     = "users"
	documentCollection = "documents"
)

func init() {
	envconfig.MustProcess("", &cfg)
	zerolog.SetGlobalLevel(zerolog.InfoLevel)
	zerolog.LevelFieldName = "severity"
}

func main() {
	ctx := context.Background()

	// Only run OpenTelemetry if not in local mode
	if !cfg.Local {
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
	}

	firestoreClient, err := firestore.NewClient(ctx, cfg.ProjectID)
	if err != nil {
		log.Fatal().Msgf("Failed to create Firestore client: %v", err)
	}

	storageClient, err := storage.NewClient(ctx)
	if err != nil {
		log.Fatal().Msgf("Failed to create GCS client: %v", err)
	}

	documentDataStore := db.NewFirestoreRepository[models.Document](firestoreClient, documentCollection)
	userDatastore := db.NewFirestoreRepository[models.User](firestoreClient, userCollection)
	storageStore, err := gcs.NewGCSStorage(storageClient, cfg.BucketName)
	if err != nil {
		log.Fatal().Msgf("Failed to create GCS storage client: %v", err)
	}

	documentService := services.NewDocumentService(storageStore, documentDataStore)
	documentHandler := handlers.NewDocumentHandler(documentService)

	userService := services.NewUserService(userDatastore)
	userHandler := handlers.NewUserHandler(userService)

	r := router.NewRouter(cfg.ServiceName, cfg.Local, &cfg.Port)

	documentHandler.RegisterRoutes(r.Engine)
	userHandler.RegisterRoutes(r.Engine)

	log.Fatal().Err(r.Run()).Msg("Failed to run server")
}
