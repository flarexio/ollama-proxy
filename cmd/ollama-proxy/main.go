package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/nats-io/nats.go"
	"github.com/urfave/cli/v2"
	"go.uber.org/zap"

	proxy "github.com/flarexio/ollama-proxy"
)

func main() {
	app := &cli.App{
		Name: "ollama-proxy",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "ollama-url",
				EnvVars: []string{"OLLAMA_SERVICE_URL"},
				Value:   "http://localhost:11434",
			},
			&cli.StringFlag{
				Name:    "nats",
				EnvVars: []string{"NATS_URL"},
				Value:   "wss://nats.flarex.io",
			},
			&cli.StringFlag{
				Name:    "creds",
				EnvVars: []string{"NATS_CREDS"},
			},
		},
		Action: run,
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}

func run(cli *cli.Context) error {
	log, err := zap.NewDevelopment()
	if err != nil {
		return err
	}
	defer log.Sync()

	zap.ReplaceGlobals(log)

	instance := cli.String("ollama-url")

	svc, err := proxy.NewService(instance)
	if err != nil {
		return err
	}

	svc = proxy.LoggingMiddleware(log)(svc)

	ctx := context.Background()
	ver, err := svc.Version(ctx)
	if err != nil {
		return err
	}

	log.Info("ollama connected",
		zap.String("service", "ollama"),
		zap.String("version", ver),
	)

	natsURL := cli.String("nats")
	natsCreds := cli.String("creds")

	nc, err := nats.Connect(natsURL,
		nats.Name("ollama-proxy"),
		nats.UserCredentials(natsCreds),
	)
	if err != nil {
		return err
	}
	defer nc.Drain()

	nc.Subscribe("ollama.version", proxy.VersionHandler(svc))
	nc.Subscribe("ollama.models", proxy.ListHandler(svc))
	nc.Subscribe("ollama.chat", proxy.ChatHandler(svc))

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	sign := <-quit // Wait for a termination signal

	log.Info("graceful shutdown", zap.String("singal", sign.String()))
	return nil
}
