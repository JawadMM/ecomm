package main

import (
	"context"
	"log"
	"time"

	"github.com/JawadMM/ecomm/catalog"
	"github.com/JawadMM/ecomm/events"
	"github.com/JawadMM/ecomm/order"
	"github.com/kelseyhightower/envconfig"
	"github.com/tinrab/retry"
)

type config struct {
	DatabaseURL       string `envconfig:"DATABASE_URL"`
	NatsURL           string `envconfig:"NATS_URL"`
	CatalogServiceURL string `envconfig:"CATALOG_SERVICE_URL"`
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

	var sc order.StockClient
	if cfg.CatalogServiceURL != "" {
		var catalogClient *catalog.Client
		retry.ForeverSleep(2*time.Second, func(_ int) (err error) {
			catalogClient, err = catalog.NewClient(cfg.CatalogServiceURL)
			if err != nil {
				log.Printf("Failed to connect to catalog service: %v. Retrying...", err)
			}
			return
		})
		defer catalogClient.Close()
		sc = catalogClient
		log.Println("Connected to catalog service")
	}

	log.Println("Listening on port 8080...")
	s := order.NewService(r, pub, sc)
	log.Fatal(order.ListenGRPC(s, 8080))
}
