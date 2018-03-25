package main

import (
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/improbable-eng/grpc-web/go/grpcweb"
	"github.com/jawher/mow.cli"
	"github.com/sbramin/grpc-demo/cmd/third-party-service/pkg/pb/tps"
	"github.com/sbramin/grpc-demo/internal/service"
	"github.com/sbramin/grpc-demo/pkg/pb/example"

	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	"github.com/grpc-ecosystem/go-grpc-middleware"
	"github.com/grpc-ecosystem/go-grpc-middleware/logging/logrus"
	"github.com/grpc-ecosystem/go-grpc-middleware/recovery"
	"github.com/grpc-ecosystem/go-grpc-prometheus"
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
		Value:  8090,
		EnvVar: "GRPC_PORT",
	})
	grpcWeb := app.Int(cli.IntOpt{
		Name:   "grpc-web",
		Desc:   "GRPC Web",
		Value:  9090,
		EnvVar: "GRPC_WEB",
	})
	tpsAPI := app.String(cli.StringOpt{
		Name:   "tps-api",
		Desc:   "Third-Party-Service-API",
		Value:  "localhost:8091",
		EnvVar: "TPS_API",
	})

	app.Action = func() {
		grpc_prometheus.EnableHandlingTimeHistogram()
		logger := setUpLogger(*logLevel)
		grpc_logrus.ReplaceGrpcLogger(logger)

		db := "Greetings "

		tpsConn, err := grpc.Dial(*tpsAPI, grpc.WithInsecure(), grpc.WithTimeout(1*time.Second))
		if err != nil {
			log.Panic("conn err", err)
		}
		tpsCli := tps.NewThirdPartyServiceClient(tpsConn)
		defer tpsConn.Close()

		svc := service.New(db, tpsCli)

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

		example.RegisterGreeterServer(gSrv, svc)

		go waitForShutdown(func() {
			logger.Warn("shutdown")
			gSrv.GracefulStop()
		})

		go func() {
			wrappedServer := grpcweb.WrapServer(gSrv)
			handler := func(resp http.ResponseWriter, req *http.Request) {
				wrappedServer.ServeHTTP(resp, req)
			}
			httpServer := http.Server{
				Addr:    fmt.Sprintf(":%d", *grpcWeb),
				Handler: http.HandlerFunc(handler),
			}
			if err := httpServer.ListenAndServe(); err != nil {
				log.Panicf("failed starting http server: %v", err)
			}
		}()

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
