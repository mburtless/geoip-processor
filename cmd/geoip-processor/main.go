package main

import (
	"context"
	"errors"
	"os"
	"os/signal"
	"strconv"

	"go.uber.org/zap"
)

type config struct {
	addr          string
	maxConStreams int
	logger        *zap.Logger
}

func main() {
	cfg := getConfig()

	// init logger
	cfg.logger, _ = zap.NewDevelopment()

	// run
	ctx, cancel := context.WithCancel(context.Background())
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	defer func() {
		signal.Stop(c)
		cancel()
	}()

	go func() {
		select {
		case <-c:
			cancel()
		case <-ctx.Done():
		}
	}()

	err := run(ctx, cfg)
	if err != nil && !errors.Is(err, context.Canceled) {
		cfg.logger.Fatal("err ", zap.Error(err))
	}
}

func run(ctx context.Context, cfg *config) error {
	return errors.New("NYI")
}

func getConfig() *config {
	var (
		ok  bool
		cfg config
		err error
	)
	// parse addr, max concurrent streams
	cfg.addr, ok = os.LookupEnv("ADDR")
	if !ok {
		cfg.addr = "localhost:8000"
	}
	cfg.maxConStreams, err = strconv.Atoi(os.Getenv("MAX_CONCURRENT_STREAMS"))
	if err != nil {
		cfg.maxConStreams = 1000
	}
	return &cfg
}
