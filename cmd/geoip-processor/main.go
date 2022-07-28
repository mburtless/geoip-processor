package main

import (
	"context"
	"errors"
	"net"
	"os"
	"os/signal"
	"strconv"

	"github.com/mburtless/geoip-processor/internal/server"
	"go.uber.org/zap"
	"google.golang.org/grpc"
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
		cfg.logger.Fatal("error running server", zap.Error(err))
	}
}

func run(ctx context.Context, cfg *config) error {
	lis, err := net.Listen("tcp", cfg.addr)
	if err != nil {
		return err
	}

	opts := []grpc.ServerOption{grpc.MaxConcurrentStreams(uint32(cfg.maxConStreams))}
	s := grpc.NewServer(opts...)

	srv := server.NewServer(cfg.logger)
	srv.RegisterServer(s)

	errChan := make(chan error, 1)
	go func() {
		errChan <- s.Serve(lis)
	}()

	cfg.logger.Info("starting server", zap.String("address", cfg.addr))
	select {
	case err := <-errChan:
		return err
	case <-ctx.Done():
		cfg.logger.Info("stopping server")
		s.GracefulStop()
		return ctx.Err()
	}
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
