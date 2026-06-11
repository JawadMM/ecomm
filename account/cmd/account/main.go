package main

import (
	"log"
	"time"

	"github.com/JawadMM/ecomm/account"
	"github.com/kelseyhightower/envconfig"
	"github.com/tinrab/retry"
)

type config struct {
	DatabaseURL string `env:"DATABASE_URL"`
}

func main() {
	var cfg config
	err := envconfig.Process("", &cfg)
	if err != nil {
		log.Fatal(err)
	}

	var r account.Repository
	retry.ForeverSleep(2*time.Second, func(_ int) (err error) {
		r, err = account.NewPostgresRepository(cfg.DatabaseURL)
		if err != nil {
			log.Printf("Failed to connect to database: %v. Retrying...", err)
		}
		return
	})
	defer r.Close()

	log.Println("Listening on port 8080...")
	s := account.NewService(r)
	log.Fatal(account.ListenGRPC(s, 8080))
}
