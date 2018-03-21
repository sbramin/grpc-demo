package main

import (
	"fmt"
	"net"
	"os"
	"os/signal"
	"syscall"

	"github.com/jawher/mow.cli"
	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	"github.com/grpc-ecosystem/go-grpc-middleware"
	"github.com/grpc-ecosystem/go-grpc-middleware/logging/logrus"
	"github.com/grpc-ecosystem/go-grpc-middleware/recovery"
	"github.com/grpc-ecosystem/go-grpc-prometheus"

	"github.com/sbramin/grpc-demo/cmd/third-party-service/internal/service"
	"github.com/sbramin/grpc-demo/cmd/third-party-service/pkg/pb/tps"
)

const (
	appName     = "grpc-demo"
	appDesc     = "GRPC Demo"
	appOwner    = "sbramin"
	appOwnerURL = "http://sbramin.com"
	appURL      = "https://github.com/sbramin/grpc-demo"
)

var (
	gitHash string
)

func main() {
	app := cli.App(appName, appDesc)

	logLevel := app.String(cli.StringOpt{
		Name:   "log-level",
		Desc:   "log level [debug|info|warn|error]",
		EnvVar: "LOG_LEVEL",
		Value:  "info",
	})
	grpcPort := app.Int(cli.IntOpt{
		Name:   "grpc-port",
		Desc:   "GRPC port",
		Value:  8091,
		EnvVar: "GRPC_PORT",
	})

	app.Action = func() {
		grpc_prometheus.EnableClientHandlingTimeHistogram()
		logger := setUpLogger(*logLevel)
		grpc_logrus.ReplaceGrpcLogger(logger)

		svc := service.New()

		lis, err := net.Listen("tcp", fmt.Sprintf(":%d", *grpcPort))
		if err != nil {
			log.Panicf("failed to listen: %v", err)
		}
		gSrv := grpc.NewServer(
			grpc.UnaryInterceptor(grpc_middleware.ChainUnaryServer(
				grpc_prometheus.UnaryServerInterceptor,
				grpc_recovery.UnaryServerInterceptor(),
				grpc_logrus.UnaryServerInterceptor(logger),
			)),
		)
		tps.RegisterThirdPartyServiceServer(gSrv, svc)
		go waitForShutdown(func() {
			logger.Warn("shutdown")
			gSrv.GracefulStop()
		})
		reflection.Register(gSrv)
		if err := gSrv.Serve(lis); err != nil {
			log.Panicf("failed to serve: %v", err)
		}
	}
	log.Info("app starting")
	if err := app.Run(os.Args); err != nil {
		log.WithError(err).Panic("app stopped with error")
	}
}

func setUpLogger(level string) *log.Entry {
	l, err := log.ParseLevel(level)
	if err != nil {
		log.WithError(err).Panic("error parsing log level")
	}
	logger := log.Logger{
		Out:       os.Stderr,
		Formatter: &log.JSONFormatter{},
		Hooks:     make(log.LevelHooks),
		Level:     l,
	}
	return log.NewEntry(&logger)
}

func waitForShutdown(shutdown func()) {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	<-sigChan
	shutdown()
}
