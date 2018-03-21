package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/grpc-ecosystem/go-grpc-middleware/logging/logrus"
	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"github.com/jawher/mow.cli"
	"github.com/sbramin/grpc-demo/pkg/pb/example"
	log "github.com/sirupsen/logrus"
	"github.com/utilitywarehouse/swagger-ui/swaggerui"
	"google.golang.org/grpc"
)

const (
	appName        = "grpc-demo-rest-proxy"
	appDescription = "Provides HTTP/JSON API for the grpc-demo-rest-proxy"
	appNamespace   = "demo"
)

var (
	gitHash string
)

func main() {
	app := cli.App(appName, appDescription)

	appLogLevel := app.String(cli.StringOpt{
		Name:   "log-Level",
		Desc:   "log level [debug|info|warn|error]",
		EnvVar: "LOG_LEVEL",
		Value:  "info",
	})

	grpcPort := app.Int(cli.IntOpt{
		Name:   "grpc-port",
		Desc:   "GRPC port",
		Value:  8090,
		EnvVar: "GRPC_PORT",
	})

	httpPort := app.Int(cli.IntOpt{
		Name:   "http-port",
		Desc:   "HTTP port",
		Value:  8080,
		EnvVar: "HTTP_PORT",
	})

	app.Action = func() {
		logger := setUpLogger(*appLogLevel)
		grpc_logrus.ReplaceGrpcLogger(logger)

		mux := http.NewServeMux()
		exampleMux := runtime.NewServeMux()

		mux.Handle("/", exampleMux)
		mux.Handle("/swagger-ui/", swaggerui.UIHandler())
		mux.Handle("/swagger.json", swaggerui.FileHandler())

		ctx := context.Background()
		ctx, cancel := context.WithCancel(ctx)
		defer cancel()
		dialOpts := []grpc.DialOption{grpc.WithInsecure()}
		err := example.RegisterExampleHandlerFromEndpoint(ctx, exampleMux, fmt.Sprintf("localhost:%d", *grpcPort), dialOpts)
		if err != nil {
			logger.WithError(err).Panic("cannot register handler")
		}

		server := http.Server{
			Addr:    fmt.Sprintf(":%d", *httpPort),
			Handler: mux,
		}

		go waitForShutdown(func() {
			logger.Warn("Shutdown")
			server.Shutdown(context.Background())
		})

		if err := server.ListenAndServe(); err != nil {
			logger.WithError(err).Panicf("cannot listen and serve on port: %d", *httpPort)
		}
	}

	if err := app.Run(os.Args); err != nil {
		log.WithError(err).Panic("App stopped with error")
	}
}

func waitForShutdown(shutdown func()) {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	<-sigChan
	shutdown()
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
