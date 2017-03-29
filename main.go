package main

import (
	"context"
	"crypto/tls"
	"net"
	"os"
	"os/signal"
	"syscall"

	"golang.org/x/sync/errgroup"

	"go.uber.org/zap"

	"github.com/openfresh/plasma/config"
	"github.com/openfresh/plasma/log"
	"github.com/openfresh/plasma/protobuf"
	"github.com/openfresh/plasma/pubsub"
	"github.com/openfresh/plasma/server"
	"github.com/openfresh/plasma/subscriber"
	"github.com/soheilhy/cmux"
	"google.golang.org/grpc"
)

type Service struct {
	Serve    func(l net.Listener) error
	Listener net.Listener
}

func plasmaListener(logger *zap.Logger, config config.Config) net.Listener {
	l, err := net.Listen("tcp", ":"+config.Port)
	if err != nil {
		logger.Fatal("failed to listen",
			zap.Error(err),
			zap.String("port", config.Port),
		)
	}

	if config.TLS.CertFile != "" && config.TLS.KeyFile != "" {
		cer, err := tls.LoadX509KeyPair(config.TLS.CertFile, config.TLS.KeyFile)
		if err != nil {
			logger.Fatal("failed to load TLS credentials for TCP",
				zap.Error(err),
				zap.String("certFile", config.TLS.CertFile),
				zap.String("keyFile", config.TLS.KeyFile),
			)
		}

		logger.Info("enable TLS mode",
			zap.String("certFile", config.TLS.CertFile),
			zap.String("keyFile", config.TLS.KeyFile),
		)
		return tls.NewListener(l, &tls.Config{
			Certificates: []tls.Certificate{cer},
		})
	}

	logger.Info("non TLS mode")
	return l
}

func main() {
	config := config.New()

	accessLogger, err := log.NewLogger(config.AccessLog)
	if err != nil {
		panic(err)
	}
	errorLogger, err := log.NewLogger(config.ErrorLog)
	if err != nil {
		panic(err)
	}

	l := plasmaListener(errorLogger, config)
	defer l.Close()

	pubsuber := pubsub.NewPubSub()

	sub, err := subscriber.New(pubsuber, errorLogger, config)
	if err != nil {
		errorLogger.Fatal("failed to create subscriber",
			zap.Error(err),
			zap.String("type", config.Subscriber.Type),
			zap.Duration("mockDuration", config.Subscriber.Mock.Interval),
			zap.Object("redis", config.Subscriber.Redis),
		)
	}
	go func() {
		sub := sub
		if err := sub.Subscribe(); err != nil {
			errorLogger.Fatal("failed to subscribe",
				zap.String("type", config.Subscriber.Type),
				zap.Object("redis", config.Subscriber.Redis),
				zap.Error(err),
			)
		}
	}()

	// For Native Client
	grpcServerOption := server.Option{
		PubSuber:     pubsuber,
		AccessLogger: accessLogger,
		ErrorLogger:  errorLogger,
		Config:       config,
	}
	grpcServer := grpc.NewServer()
	proto.RegisterStreamServiceServer(grpcServer, server.NewStreamServer(grpcServerOption))

	// For Web Front End
	sseServerOption := server.Option{
		PubSuber:     pubsuber,
		AccessLogger: accessLogger,
		ErrorLogger:  errorLogger,
		Config:       config,
	}
	sseServer := server.NewSSEServer(sseServerOption)

	// For Meta (HealthCheck, Metrics)
	metaServer := server.NewMetaServer(server.Option{
		AccessLogger: accessLogger,
		ErrorLogger:  errorLogger,
		Config:       config,
	})

	// for graceful shutdown
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh,
		syscall.SIGHUP,
		syscall.SIGINT,
		syscall.SIGKILL,
		syscall.SIGQUIT,
		syscall.SIGTERM)
	go func() {
		<-sigCh

		eg := errgroup.Group{}
		eg.Go(func() error {
			errorLogger.Info("shutdown gRPC Server gracefully...")
			grpcServer.GracefulStop()
			return nil
		})
		eg.Go(func() error {
			errorLogger.Info("shutdown sseServer gracefully...")
			return sseServer.Shutdown(context.Background())
		})
		eg.Go(func() error {
			errorLogger.Info("shutdown metaServer gracefully...")
			return metaServer.Shutdown(context.Background())
		})
		if err := eg.Wait(); err != nil {
			opErr, ok := err.(*net.OpError)

			// NOTE: Ignore errors that occur when closing the file descriptor because it is an assumed error.
			if ok && opErr.Op == "close" {
				return
			}
			errorLogger.Fatal("failed to shutdown gracefully",
				zap.Error(err),
			)
		}
	}()

	m := cmux.New(l)
	services := []Service{
		{
			Serve:    grpcServer.Serve,
			Listener: m.MatchWithWriters(cmux.HTTP2MatchHeaderFieldSendSettings("content-type", "application/grpc")),
		},
		{
			Serve:    sseServer.Serve,
			Listener: m.Match(cmux.HTTP1HeaderField("Accept", "text/event-stream")),
		},
		{
			Serve:    metaServer.Serve,
			Listener: m.Match(cmux.HTTP1()),
		},
	}

	for _, service := range services {
		go service.Serve(service.Listener)
	}

	m.Serve()
}
