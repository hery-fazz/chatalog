package main

import (
	"context"
	"fmt"

	"github.com/defryfazz/fazztalog/config"
	"github.com/defryfazz/fazztalog/internal/app"
)

func main() {

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	client, err := setupWhatsmeowClient(ctx)
	if err != nil {
		panic(fmt.Sprintf("failed to setup whatsmeow client: %v", err))
	}

	db, err := setupSQLiteDatabase(config.SQLitePath)
	if err != nil {
		panic(fmt.Sprintf("failed to setup sqlite database: %v", err))
	}

	appContainer := app.SetupApp(app.SetupAppParams{
		OpenAIToken:   config.OpenAIToken,
		DB:            db,
		TempDirectory: config.TempFolderPath,
	})
	eventHandler := &EventHandler{
		client:       client,
		appContainer: appContainer,
	}
	client.AddEventHandler(eventHandler.Handle(ctx))

	connectWhatsmeowClient(client)

}
