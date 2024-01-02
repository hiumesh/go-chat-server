package main

import (
	"context"
	"os/signal"
	"syscall"

	"github.com/hiumesh/go-chat-server/cmd"
	"github.com/sirupsen/logrus"
)

func main() {
	// err := godotenv.Load(".env")
	// if err != nil {
	// 	fmt.Printf("Could not load the .env file.")
	// 	os.Exit(1)
	// }

	// PORT := os.Getenv("PORT")
	// if PORT == "" {
	// 	fmt.Printf("PORT environment not found.")
	// 	os.Exit(1)
	// }

	// api := api.SetupAPI()
	// log.Fatal(api.Run(":" + PORT))

	execCtx, execCancel := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGHUP, syscall.SIGINT)
	defer execCancel()

	go func() {
		<-execCtx.Done()
		logrus.Info("received graceful shutdown signal")
	}()

	if err := cmd.RootCommand().ExecuteContext(execCtx); err != nil {
		logrus.WithError(err).Fatal(err)
	}
}
