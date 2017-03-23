package main

import (
	"crypto/tls"
	"net"
	"net/http"

	"go.uber.org/zap"

	"bufio"
	"io"

	"strings"

	"github.com/openfresh/plasma/config"
	"github.com/openfresh/plasma/log"
	"github.com/openfresh/plasma/protobuf"
	"github.com/openfresh/plasma/pubsub"
	"github.com/openfresh/plasma/server"
	"github.com/openfresh/plasma/subscriber"
	"github.com/soheilhy/cmux"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

type Service struct {
	Serve    func(l net.Listener) error
	Matchers []cmux.Matcher
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
	if config.TLS.CertFile != "" && config.TLS.KeyFile != "" {
		cer, err := tls.LoadX509KeyPair(config.TLS.CertFile, config.TLS.KeyFile)
		if err != nil {
			logger.Fatal("failed to load TLS credentials for TCP",
				zap.Error(err),
				zap.String("certFile", config.TLS.CertFile),
				zap.String("keyFile", config.TLS.KeyFile),
			)
		}

		tlsConfig := &tls.Config{
			Certificates: []tls.Certificate{cer},
		}

		l, err := tls.Listen("tcp", ":"+config.Port, tlsConfig)
		if err != nil {
			logger.Fatal("failed to listen TLS",
				zap.Error(err),
				zap.String("certFile", config.TLS.CertFile),
				zap.String("keyFile", config.TLS.KeyFile),
			)
		}
		logger.Info("enable TLS mode",
			zap.String("certFile", config.TLS.CertFile),
			zap.String("keyFile", config.TLS.KeyFile),
		)
		return l
	} else {
		l, err := net.Listen("tcp", ":"+config.Port)
		if err != nil {
			logger.Fatal("failed to listen",
				zap.Error(err),
				zap.String("port", config.Port),
			)
		}
		logger.Info("non TLS mode")
		return l
	}
}

func plasmaGRPCServerOption(logger *zap.Logger, certFile string, keyFile string) []grpc.ServerOption {
	var opts []grpc.ServerOption
	if certFile != "" && keyFile != "" {
		creds, err := credentials.NewServerTLSFromFile(certFile, keyFile)
		if err != nil {
			logger.Fatal("failed to load TLS credentials for gRPC",
				zap.Error(err),
				zap.String("certFile", certFile),
				zap.String("keyFile", certFile),
			)
		}
		opts = []grpc.ServerOption{grpc.Creds(creds)}
	}
	return opts
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
	opts := plasmaGRPCServerOption(errorLogger, config.TLS.CertFile, config.TLS.KeyFile)
	grpcServer := grpc.NewServer(opts...)
	proto.RegisterStreamServiceServer(grpcServer, server.NewStreamServer(pubsuber, accessLogger, errorLogger, config))

	// For Web Front End
	sseServer := server.NewSSEServer(pubsuber, accessLogger, errorLogger, config)

	// For AWS ELB
	healthCheckServer := server.NewHealthCheckServer(accessLogger, errorLogger, config)

	services := []Service{
		{
			Serve: grpcServer.Serve,
			Matchers: []cmux.Matcher{
				cmux.HTTP2HeaderField("content-type", "application/grpc"),
			},
		},
		{
			Serve: healthCheckServer.Serve,
			Matchers: []cmux.Matcher{
				ELBHealthCheckMatcher(),
			},
		},
		{
			Serve: sseServer.Serve,
			Matchers: []cmux.Matcher{
				cmux.HTTP1(),
			},
		},
	}

	m := cmux.New(l)
	for _, service := range services {
		go service.Serve(m.Match(service.Matchers...))
	}

	m.Serve()
}
