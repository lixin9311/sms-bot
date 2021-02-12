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
	"github.com/maltegrosse/go-modemmanager"

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
		log.Fatal("an user id must be provided")
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

func openModem() (modemmanager.Modem, modemmanager.ModemMessaging, error) {
	mmgr, err := modemmanager.NewModemManager()
	if err != nil {
		return nil, nil, err
	}
	err = mmgr.ScanDevices()
	if err != nil {
		return nil, nil, err
	}
	modems, err := mmgr.GetModems()
	if err != nil {
		return nil, nil, err
	}
	for _, modem := range modems {
		messaging, err := modem.GetMessaging()
		if err != nil {
			return nil, nil, err
		}
		return modem, messaging, nil
	}
	return nil, nil, fmt.Errorf("modem not found")
}

func main() {
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	modem, msg, err := openModem()
	if err != nil {
		log.Fatalf("failed to open modem: %v", err)
	}
	fmt.Println("openting database")
	client, err := openSQLite()
	if err != nil {
		log.Fatalf("failed to open db: %v", err)
	}
	defer client.Close()
	man := manager.NewManager(modem, msg, client)
	defer man.UnsubscribeSMS()
	defer man.UnsubscribeCall()
	_, err = man.LoadSMS(ctx)
	if err != nil {
		log.Fatalf("failed to load current sms: %v", err)
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

	quit := make(chan os.Signal)
	signal.Notify(quit, syscall.SIGTERM, syscall.SIGINT)
	<-quit
}
