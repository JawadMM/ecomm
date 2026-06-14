package main

import (
	"context"
	"log"
	"time"

	"github.com/JawadMM/ecomm/events"
	"github.com/JawadMM/ecomm/order"
	"github.com/kelseyhightower/envconfig"
	"github.com/tinrab/retry"
)

type config struct {
	DatabaseURL string `envconfig:"DATABASE_URL"`
	NatsURL     string `envconfig:"NATS_URL"`
}

func main() {
	var cfg config
	err := envconfig.Process("", &cfg)
	if err != nil {
		log.Fatal(err)
	}

	var r order.Repository
	retry.ForeverSleep(2*time.Second, func(_ int) (err error) {
		r, err = order.NewMongoRepository(cfg.DatabaseURL)
		if err != nil {
			log.Printf("Failed to connect to database: %v. Retrying...", err)
		}
		return
	})
	ctx := context.Background()
	defer r.Close(ctx)

	var pub *events.Publisher
	if cfg.NatsURL != "" {
		pub, err = events.NewPublisher(cfg.NatsURL)
		if err != nil {
			log.Fatalf("Failed to connect to NATS: %v", err)
		}
		defer pub.Close()
		log.Println("Connected to NATS")
	}

	log.Println("Listening on port 8080...")
	s := order.NewService(r, pub)
	log.Fatal(order.ListenGRPC(s, 8080))
}
