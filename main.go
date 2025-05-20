package main

import (
	"fmt"
	"go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/store/sqlstore"
	"os"
	"strings"
	"time"

	_ "github.com/mattn/go-sqlite3"
	waLog "go.mau.fi/whatsmeow/util/log"
)

func main() {
	NewBot("6281244996097", func(msg string) {
		println(msg)
	})

}

func NewBot(id string, callback func(string)) *whatsmeow.Client {
	if id == "" {
		callback("Nomor ?")
		return nil
	}
	id = strings.ReplaceAll(id, "admin", "")

	dbLog := waLog.Stdout("Database", "INFO", true)

	container, err := sqlstore.New(context.Background(), "sqlite3", "file:"+id+".db?_foreign_keys=on", dbLog)
	if err != nil {
		callback("Kesalahan (error)\n" + fmt.Sprintf("%s", err))
		return nil
	}
	deviceStore, err := container.GetFirstDevice(context.Background())
	if err != nil {
		callback("Kesalahan (error)\n" + fmt.Sprintf("%s", err))
		return nil
	}
	clientLog := waLog.Stdout("Client", "INFO", true)
	client := whatsmeow.NewClient(deviceStore, clientLog)

	err = client.Connect()
	if err != nil {
		callback("Kesalahan (error)\n" + fmt.Sprintf("%s", err))
		return nil
	}

	// Timer loop to restart every 5 minutes
	startTime := time.Now()
	ticker := time.NewTicker(1 * time.Second)

	for range ticker.C {
		// Check if 5 minutes have passed
		if time.Since(startTime) > 2*time.Minute {
			ticker.Stop()       // Stop the current ticker
			client.Disconnect() // Disconnect the client
			fmt.Println("Restarting...")
			//time.Sleep(2 * time.Second) // Optional delay before restart
			NewBot(id, callback) // Restart the bot
			return nil
		}

		// Pair phone logic
		if client.Store.ID == nil {
			client.PairPhone(context.Background(), id, true, whatsmeow.PairClientChrome, "Chrome (Linux)")
			os.Remove(id + ".db")
		}
	}

	return client
}
