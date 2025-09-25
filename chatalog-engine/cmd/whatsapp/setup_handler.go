package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/defryfazz/fazztalog/config"
	"github.com/defryfazz/fazztalog/internal/app"
	"github.com/google/uuid"
	"github.com/openai/openai-go"
	"go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/proto/waE2E"
	"go.mau.fi/whatsmeow/types/events"
	"google.golang.org/protobuf/proto"
)

var errSenderNotAuthenticated = fmt.Errorf("sender not authenticated")

type EventHandler struct {
	client       *whatsmeow.Client
	appContainer app.AppContainer
}

func (h *EventHandler) Handle(ctx context.Context) whatsmeow.EventHandler {
	return func(evt any) {
		switch v := evt.(type) {
		case *events.Message:
			err := h.authenticateSender(v)
			if err != nil {
				return
			}

			textMessage := ""
			switch {
			case getMessage(v) != "":
				if !hasActualText(v) {
					return
				}
				textMessage = getMessage(v)
			case v.Message.GetAudioMessage() != nil:
				audioMessage := v.Message.GetAudioMessage()

				audioFileName := fmt.Sprintf("%s/transcriptions/%s.wav", config.TempFolderPath, uuid.New().String())
				if err := os.MkdirAll(fmt.Sprintf("%s/transcriptions", config.TempFolderPath), 0755); err != nil {
					log.Printf("error creating directory: %v\n", err)
					return
				}
				f, _ := os.Create(audioFileName)
				err := h.client.DownloadToFile(ctx, audioMessage, f)
				if err != nil {
					log.Printf("error downloading audio: %v\n", err)
					return
				}
				f.Close()
				defer os.Remove(audioFileName)

				ff, err := os.Open(audioFileName)
				if err != nil {
					log.Printf("error opening audio: %v\n", err)
					return
				}
				defer ff.Close()

				res, err := h.appContainer.OpenAIClient.Audio.Transcriptions.New(ctx, openai.AudioTranscriptionNewParams{
					Model: openai.AudioModelWhisper1,
					File:  ff,
				})
				if err != nil {
					log.Printf("error transcripting audio: %v\n", err)
					return
				}

				textMessage = res.Text
			}

			if textMessage == "" {
				return
			}

			// todo: Process message here (generate image)

			_, err = h.client.SendMessage(ctx, v.Info.Chat, &waE2E.Message{
				Conversation: proto.String("Hi, you said: " + textMessage),
			})
			if err != nil {
				log.Printf("error sending response message: %v\n", err)
				return
			}
		}
	}
}

func getMessage(evt *events.Message) string {
	if evt.Message.GetConversation() != "" {
		return evt.Message.GetConversation()
	}
	if evt.Message.GetExtendedTextMessage() != nil && evt.Message.GetExtendedTextMessage().Text != nil {
		return evt.Message.GetExtendedTextMessage().GetText()
	}
	return ""
}

func hasActualText(evt *events.Message) bool {
	text := getMessage(evt)
	if text == "" {
		return false
	}
	ext := evt.Message.GetExtendedTextMessage()
	if ext != nil && ext.GetContextInfo() != nil {
		for _, jid := range ext.GetContextInfo().GetMentionedJID() {
			if idx := strings.Index(jid, "@"); idx > 0 {
				number := jid[:idx]
				text = strings.ReplaceAll(text, "@"+number, "")
			}
		}
	}
	return strings.TrimSpace(text) != ""
}

func (h *EventHandler) authenticateSender(evt *events.Message) error {
	if evt.Info.IsGroup {
		return errSenderNotAuthenticated
	}

	whitelistedPhones := []string{
		"6282123430340", // Defry
		"6285224416325", // Hery
		"6282148924797", // Farrel
		"6281575749888", // Nurwanto
		"6281287456169", // Alex
	}

	senderJID := evt.Info.Sender.ToNonAD().String()
	splittedJID := strings.Split(senderJID, "@")
	if len(splittedJID) < 2 {
		return fmt.Errorf("invalid sender JID: %s", senderJID)
	}
	senderPhone := splittedJID[0]
	for _, whitelistedPhone := range whitelistedPhones {
		if senderPhone == whitelistedPhone {
			return nil
		}
	}

	return errSenderNotAuthenticated
}
