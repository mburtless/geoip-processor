package main

import (
	"context"
	"errors"
	"log"
	"net"
	"os"
	"os/signal"
	"strconv"

	"github.com/mburtless/geoip-processor/internal/server"
	"github.com/oschwald/geoip2-golang"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

type config struct {
	addr          string
	maxConStreams int
	logger        *zap.Logger
	dbPath        string
}

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	cfg := getConfig()

	// init logger
	cfg.logger, _ = zap.NewDevelopment()

	err := run(ctx, cfg)
	if err != nil && !errors.Is(err, context.Canceled) {
		cfg.logger.Fatal("error running server", zap.Error(err))
	}
}

func run(ctx context.Context, cfg *config) error {
	db, err := geoip2.Open(cfg.dbPath)
	if err != nil {
		return err
	}
	defer db.Close()

	lis, err := net.Listen("tcp", cfg.addr)
	if err != nil {
		return err
	}

	opts := []grpc.ServerOption{grpc.MaxConcurrentStreams(uint32(cfg.maxConStreams))}
	s := grpc.NewServer(opts...)

	srv := server.NewServer(cfg.logger, db)
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
	cfg.dbPath, ok = os.LookupEnv("GEOIP_DB")
	if !ok || cfg.dbPath == "" {
		log.Fatal("GEOIP_DB required")
	}

	// TODO: allow config of which http header to extract req IP from (XFF, x-real-ip, etc)
	// TODO: allow config of which http header to inject country code in
	return &cfg
}
