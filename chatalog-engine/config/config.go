package config

import "sync"

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
	})
}
