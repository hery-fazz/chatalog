package config

import (
	"log"
	"sync"
)

var (
	once sync.Once

	TempFolderPath   string
	WhatsmeowSQLPath string
	SQLitePath       string

	OpenAIToken string
)

func init() {
	once.Do(func() {
		TempFolderPath = getString("TEMP_FOLDER_PATH", "")
		WhatsmeowSQLPath = getString("WHATSMEOW_SQL_PATH", "")
		SQLitePath = getString("SQLITE_PATH", "")

		OpenAIToken = getString("OPEN_AI_TOKEN", "")

		log.Println("Configuration loaded")
		log.Printf("TempFolderPath: %s\n", TempFolderPath)
		log.Printf("WhatsmeowSQLPath: %s\n", WhatsmeowSQLPath)
		log.Printf("SQLitePath: %s\n", SQLitePath)
		log.Printf("OpenAIToken: %s\n", OpenAIToken)
	})
}
