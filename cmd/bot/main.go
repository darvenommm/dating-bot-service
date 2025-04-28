package main

import (
	"log"
	"os"

	"github.com/darvenommm/dating-bot-service/internal/bot"
	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	token, exists := os.LookupEnv("TOKEN")
	if !exists {
		log.Fatalln("not found token")
	}

	bot.StartListeningBot(token)
}
