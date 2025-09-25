package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/defryfazz/fazztalog/config"
	_ "github.com/mattn/go-sqlite3"
	"go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/store"
	"go.mau.fi/whatsmeow/store/sqlstore"
	"google.golang.org/protobuf/proto"
)

func setupWhatsmeowClient(ctx context.Context) (*whatsmeow.Client, error) {
	container, err := sqlstore.New(ctx, "sqlite3", fmt.Sprintf("file:%s?_foreign_keys=on", config.WhatsmeowSQLPath), nil)
	if err != nil {
		return nil, fmt.Errorf("error initializing sqlite: %v", err)
	}
	deviceStore, err := container.GetFirstDevice(ctx)
	if err != nil {
		return nil, fmt.Errorf("error getting device: %v", err)
	}
	client := whatsmeow.NewClient(deviceStore, nil)
	return client, nil
}

func connectWhatsmeowClient(client *whatsmeow.Client) {
	store.DeviceProps.Os = proto.String("chatalog")
	if client.Store.ID == nil {
		qrChan, _ := client.GetQRChannel(context.Background())
		err := client.Connect()
		if err != nil {
			panic(err)
		}
		for evt := range qrChan {
			if evt.Event == "code" {
				log.Println("QR code:", evt.Code)
			} else {
				log.Println("Login event:", evt.Event)
			}
		}
	} else {
		err := client.Connect()
		if err != nil {
			panic(err)
		}
	}

	log.Println("WhatsApp Client has connected")
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	<-c

	log.Println("WhatsApp Client disconnected")
	client.Disconnect()
}
