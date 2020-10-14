package main

import (
	"avito/config"
	"avito/internal/api/http"
	"avito/internal/db"
	"avito/internal/service"
	"avito/internal/storage"
	"context"
	"log"
)

func main() {
	applicationConfig, err := config.ParseApplicationConfig()
	if err != nil {
		log.Fatalf("Cannot parse application config, reason: %+v", err)
	}

	emailServerConfig, err := config.ParseEmailServerConfig()
	if err != nil {
		log.Fatalf("Cannot parse email server config, reason: %+v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	pgConn, err := db.NewConnectToPG(&applicationConfig.DB, ctx)
	if err != nil {
		log.Fatalf("Cannot connect to DB, reason: %v", err)
	}

	storageAPI := storage.NewStorageAPI(pgConn)
	serviceAPI := service.NewServiceAPI(storageAPI, emailServerConfig, applicationConfig)

	go func() {
		m := service.NewMonitoring(serviceAPI)
		m.Run()
	}()

	server := http.NewServer(applicationConfig.HTTPPort, serviceAPI)
	server.Start()
	defer server.Stop()

	channel := make(chan struct{})
	<-channel
}
