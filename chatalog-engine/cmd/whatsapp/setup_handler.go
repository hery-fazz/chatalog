package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"strings"

	"github.com/defryfazz/fazztalog/config"
	"github.com/defryfazz/fazztalog/internal/ai"
	"github.com/defryfazz/fazztalog/internal/app"
	"github.com/google/uuid"
	"go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/proto/waE2E"
	"go.mau.fi/whatsmeow/types"
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
		defer func() {
			if r := recover(); r != nil {
				log.Printf("recovered from panic: %v\n", r)
			}
		}()
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

				textMessage, err = h.appContainer.AIEngine.TranscribeAudio(ctx, ff)
				if err != nil {
					log.Printf("error transcripting audio: %v\n", err)
					return
				}
			}

			if textMessage == "" {
				return
			}

			intent, err := h.appContainer.AIEngine.DetermineIntent(ctx, textMessage)
			if err != nil {
				log.Printf("error determining intent: %v\n", err)
				return
			}
			if intent.Intent != string(ai.IntentBrochureGeneration) {
				_, err = h.client.SendMessage(ctx, v.Info.Chat, &waE2E.Message{
					Conversation: proto.String("Sorry, I can't help you with that. I can only assist with brochure generation requests."),
				})
				if err != nil {
					log.Printf("error sending response message: %v\n", err)
					return
				}
				return
			}

			h.client.SendMessage(ctx, v.Info.Chat, &waE2E.Message{
				Conversation: proto.String("`Generating brochure...`"),
			})
			filePath, err := h.appContainer.MerchantService.GenerateBrochure(ctx, getPhoneFromJID(v.Info.Sender.ToNonAD().String()), intent.Products)
			if err != nil {
				log.Printf("error generating brochure: %v\n", err)
				h.client.SendMessage(ctx, v.Info.Chat, &waE2E.Message{
					Conversation: proto.String("Sorry the brochure generation failed. Please try again later."),
				})
				return
			}

			h.client.SendMessage(ctx, v.Info.Chat, &waE2E.Message{
				Conversation: proto.String("`Uploading brochure...`"),
			})
			err = h.sendImage(ctx, v.Info.Chat, filePath)
			if err != nil {
				log.Printf("error sending brochure image: %v\n", err)
				h.client.SendMessage(ctx, v.Info.Chat, &waE2E.Message{
					Conversation: proto.String("Sorry the brochure sending failed. Please try again later."),
				})
				return
			}

			// _, err = h.client.SendMessage(ctx, v.Info.Chat, &waE2E.Message{
			// 	Conversation: proto.String("Hi, you said: " + textMessage),
			// })
			// if err != nil {
			// 	log.Printf("error sending response message: %v\n", err)
			// 	return
			// }
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
	senderPhone := getPhoneFromJID(senderJID)
	if senderPhone == "" {
		return fmt.Errorf("failed to get phone from JID: %s", senderJID)
	}
	for _, whitelistedPhone := range whitelistedPhones {
		if senderPhone == whitelistedPhone {
			return nil
		}
	}

	return errSenderNotAuthenticated
}

func (h *EventHandler) sendImage(ctx context.Context, jid types.JID, filePath string) error {
	file, err := os.Open(filePath) // filePath is now filepath
	if err != nil {
		return err
	}
	defer file.Close()

	data, err := io.ReadAll(file)
	if err != nil {
		return err
	}

	upload, err := h.client.Upload(ctx, data, whatsmeow.MediaImage)
	if err != nil {
		return err
	}

	// Try to detect mimetype from file extension
	mimetype := "image/jpeg" // default
	if extIdx := strings.LastIndex(filePath, "."); extIdx != -1 {
		ext := strings.ToLower(filePath[extIdx:])
		switch ext {
		case ".png":
			mimetype = "image/png"
		case ".jpg", ".jpeg":
			mimetype = "image/jpeg"
		case ".gif":
			mimetype = "image/gif"
		}
	}

	_, err = h.client.SendMessage(ctx, jid, &waE2E.Message{
		ImageMessage: &waE2E.ImageMessage{
			URL:           proto.String(upload.URL),
			DirectPath:    proto.String(upload.DirectPath),
			MediaKey:      upload.MediaKey,
			FileLength:    proto.Uint64(uint64(len(data))),
			Mimetype:      proto.String(mimetype),
			FileEncSHA256: upload.FileEncSHA256,
			FileSHA256:    upload.FileSHA256,
		},
	})
	if err != nil {
		return err
	}

	return nil
}

func getPhoneFromJID(jid string) string {
	splittedJID := strings.Split(jid, "@")
	if len(splittedJID) < 2 {
		return ""
	}

	return splittedJID[0]
}
