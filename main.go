package main

import (
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"

	"github.com/hiumesh/go-chat-server/internal/api"
)

func main() {
	err := godotenv.Load(".env")
	if err != nil {
		fmt.Printf("Could not load the .env file.")
		os.Exit(1)
	}

	PORT := os.Getenv("PORT")
	if PORT == "" {
		fmt.Printf("PORT environment not found.")
		os.Exit(1)
	}

	api := api.SetupAPI()
	log.Fatal(api.Run(PORT))

}
