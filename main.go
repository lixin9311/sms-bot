package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/lixin9311/sms-bot/bot"
	"github.com/lixin9311/sms-bot/ent"
	"github.com/lixin9311/sms-bot/ent/migrate"
	"github.com/lixin9311/sms-bot/manager"

	_ "github.com/mattn/go-sqlite3"
)

var (
	userID   = flag.Int("user-id", 0, "user id of yourself, the bot will only responds to you")
	botToken = flag.String("token", "", "telegram bot token")
)

func init() {
	flag.Parse()
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	if *userID == 0 {
		log.Println("bot will response to anyone")
	} else if *botToken == "" {
		log.Fatal("bot token must be provided")
	}
}

func openSQLite() (client *ent.Client, err error) {
	if client, err = ent.Open("sqlite3", "file:sms.db?mode=rwc&cache=shared&_fk=1"); err != nil {
		return
	}
	// Run the auto migration tool.
	err = client.Schema.Create(
		context.Background(),
		migrate.WithDropIndex(true),
		migrate.WithDropColumn(true),
	)
	return client, err
}

func main() {
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	fmt.Println("opening database")
	client, err := openSQLite()
	if err != nil {
		log.Fatalf("failed to open db: %v", err)
	}
	defer client.Close()
	man, err := manager.NewManager(ctx, client)
	if err != nil {
		log.Fatalf("failed to init manager: %v", err)
	}
	bot, err := bot.NewBot(*botToken, *userID, man)
	if err != nil {
		log.Fatal(err)
	}

	go func() {
		if err := bot.Start(ctx); err != nil {
			log.Fatalf("failed to start bot: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGTERM, syscall.SIGINT)
	<-quit
}
