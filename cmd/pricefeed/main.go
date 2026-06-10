package main

import (
	"context"
	"flag"
	"os"
	"os/signal"
	"syscall"

	"github.com/steemfans/go-pricefeed/internal/config"
	"github.com/steemfans/go-pricefeed/internal/log"
	"github.com/steemfans/go-pricefeed/internal/price"
	"github.com/steemfans/go-pricefeed/internal/publisher"
	"github.com/steemfans/go-pricefeed/internal/scheduler"
)

func main() {
	configPath := flag.String("config", "config.json", "path to config file")
	flag.Parse()

	cfg, err := config.Load(*configPath)
	if err != nil {
		log.Println(err)
		os.Exit(1)
	}

	agg := price.NewAggregator(cfg)
	pub := publisher.NewPublisher(cfg)
	sched := scheduler.New(cfg, agg, pub)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigCh
		log.Println("shutdown signal received")
		cancel()
	}()

	if err := sched.Run(ctx); err != nil && err != context.Canceled {
		log.Printf("scheduler stopped: %v", err)
		os.Exit(1)
	}
}
