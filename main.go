package main

import (
	"TgLotteryBot/bot"
	"log"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	// Initialize the bot
	botInstance, err := bot.NewBot()
	if err != nil {
		log.Fatalf("Error creating bot: %v", err)
	}

	// Start the bot
	go botInstance.Start()

	// Handle graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down bot...")
	botInstance.Stop()
}
