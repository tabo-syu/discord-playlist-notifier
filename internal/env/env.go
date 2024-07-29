package env

import "os"

var (
	DB_HOST     = os.Getenv("DB_HOST")
	DB_PORT     = os.Getenv("DB_PORT")
	DB_NAME     = os.Getenv("DB_NAME")
	DB_USER     = os.Getenv("DB_USER")
	DB_PASSWORD = os.Getenv("DB_PASSWORD")
	DB_TIMEZONE = os.Getenv("DB_TIMEZONE")

	DISCORD_TOKEN = os.Getenv("DISCORD_ACCESS_TOKEN")
	YOUTUBE_TOKEN = os.Getenv("YOUTUBE_APIKEY")
)
