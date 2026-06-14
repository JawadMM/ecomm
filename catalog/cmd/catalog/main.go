package main

import (
	"log"
	"time"

	"github.com/JawadMM/ecomm/catalog"
	"github.com/JawadMM/ecomm/events"
	"github.com/kelseyhightower/envconfig"
	"github.com/tinrab/retry"
)

type Config struct {
	DatabaseURL string `envconfig:"DATABASE_URL"`
	NatsURL     string `envconfig:"NATS_URL"`
}

func main() {
	var cfg Config
	err := envconfig.Process("", &cfg)
	if err != nil {
		log.Fatal(err)
	}

	var r catalog.Respository
	retry.ForeverSleep(2*time.Second, func(_ int) (err error) {
		r, err = catalog.NewElasticRespository(cfg.DatabaseURL)
		if err != nil {
			log.Println(err)
		}
		return
	})
	defer r.Close()

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
	s := catalog.NewService(r, pub)
	log.Fatal(catalog.ListenGRPC(s, 8080))
}
