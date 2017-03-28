package main

import (
	"context"
	"crypto/tls"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"golang.org/x/sync/errgroup"

	"go.uber.org/zap"

	"bufio"
	"io"

	"strings"

	"github.com/openfresh/plasma/config"
	"github.com/openfresh/plasma/log"
	"github.com/openfresh/plasma/metrics"
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

const ELBHealthCheckUserAgent = "ELB-HealthChecker/"

func ELBHealthCheckMatcher() cmux.Matcher {
	return func(r io.Reader) bool {
		req, err := http.ReadRequest(bufio.NewReader(r))
		if err != nil {
			return false
		}
		userAgent := req.Header.Get("User-Agent")
		if strings.Contains(userAgent, ELBHealthCheckUserAgent) {
			return true
		}

		return false
	}
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
	grpcMetrics, err := metrics.New()
	if err != nil {
		errorLogger.Fatal("failed to create grpc metrics",
			zap.Error(err),
		)
	}
	grpcServerOption := server.Option{
		PubSuber:     pubsuber,
		AccessLogger: accessLogger,
		ErrorLogger:  errorLogger,
		Config:       config,
		Metrics:      grpcMetrics,
	}
	grpcServer := grpc.NewServer()
	proto.RegisterStreamServiceServer(grpcServer, server.NewStreamServer(grpcServerOption))

	// For Web Front End
	sseMetrics, err := metrics.New()
	if err != nil {
		errorLogger.Fatal("failed to create sse metrics",
			zap.Error(err),
		)
	}
	sseServerOption := server.Option{
		PubSuber:     pubsuber,
		AccessLogger: accessLogger,
		ErrorLogger:  errorLogger,
		Config:       config,
		Metrics:      sseMetrics,
	}
	sseServer := server.NewSSEServer(sseServerOption)

	// For AWS ELB
	healthCheckServer := server.NewHealthCheckServer(accessLogger, errorLogger, config)

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
			errorLogger.Info("shutdown healthCheckServer gracefully...")
			return healthCheckServer.Shutdown(context.Background())
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
			Serve:    healthCheckServer.Serve,
			Listener: m.Match(ELBHealthCheckMatcher()),
		},
		{
			Serve:    sseServer.Serve,
			Listener: m.Match(cmux.HTTP1()),
		},
	}

	for _, service := range services {
		go service.Serve(service.Listener)
	}

	m.Serve()
}
