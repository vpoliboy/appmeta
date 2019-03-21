/*
 * Copyright (c) Vijay Poliboyina 2019.
 */

package main

import (
	"context"
	"expvar"
	"flag"
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
	"github.com/vpoliboy/appmeta/pkg/config"
	"github.com/vpoliboy/appmeta/pkg/metadata"
	mhttp "github.com/vpoliboy/appmeta/pkg/metadata/http"
	"github.com/vpoliboy/appmeta/pkg/middleware"
	"net/http"
	"os"
	"os/signal"
	"time"
)

const (
	base = "/api/v1"
)

var (
	serverAddr string
	debug      bool
	confDir    string
)

func init() {
	flag.StringVar(&serverAddr, "addr", ":8080", "http server address")
	flag.BoolVar(&debug, "debug", false, "debug mode")
	flag.StringVar(&confDir, "conf", "./conf", "directory to look into for config files")
}

func main() {

	flag.Parse()

	logger := logrus.New()
	if debug {
		logger.Info("Running in debug mode")
		logger.SetLevel(logrus.DebugLevel)
	}

	var metadataServiceOpts []metadata.ServiceOption

	analyzerConfig, err := config.LoadAnalyzerConfig(confDir)
	if err == nil {
		metadataServiceOpts = append(metadataServiceOpts, metadata.WithMappings(analyzerConfig))
	}

	metadataService := metadata.NewService(logger, metadataServiceOpts...)

	router := mux.NewRouter()

	middlewareChain := middleware.Chain(
		middleware.PanicLoggerMiddleware(logger),
		middleware.InstrumentingMiddleware("appmeta"))

	router.Handle(base+"/stats", expvar.Handler())
	metadataHandler := mhttp.MakeHttpHandler(base, router, middlewareChain, metadataService, logger)
	router.Handle(base, metadataHandler)

	httpServer := http.Server{Addr: serverAddr, Handler: router}
	startAndWaitForShutdown(&httpServer, metadataService, logger)
}

func startAndWaitForShutdown(httpServer *http.Server, service metadata.Service, logger *logrus.Logger) {

	logger.Info("Starting http server at ", httpServer.Addr)

	// stop channel for the signal handling
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)

	// channel for reporting unusual server errors
	errChannel := make(chan error, 1)
	go func() {
		if err := httpServer.ListenAndServe(); err != http.ErrServerClosed {
			errChannel <- err
		}
	}()

	select {
	case <-stop:
		ctx, cancelFn := context.WithTimeout(context.Background(), time.Duration(time.Second*10))
		defer cancelFn()
		_ = service.Shutdown(ctx)
		_ = httpServer.Shutdown(ctx)

	case err := <-errChannel:
		logger.Error("http server quit unexpectedly, reason", err)
	}
	logger.Info("Server shutdown")
}
