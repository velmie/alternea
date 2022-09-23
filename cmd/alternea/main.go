package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/velmie/alternea/app"
	"github.com/velmie/alternea/bootstrap"
	"github.com/velmie/alternea/route"
)

const (
	shutdownTimeout = 5 * time.Second
	defaultLogLevel = app.InfoLevel
)

var (
	log app.Logger
)

func main() {
	logger := logrus.New()
	rootConfig, err := loadConfig()
	if err != nil {
		logger.Errorf("cannot read application configuration: %s", err)
		os.Exit(1)
	}

	var logLevel = defaultLogLevel
	if rootConfig.LogLevel != "" {
		logLevel, err = app.ParseLevel(rootConfig.LogLevel)
		if err != nil {
			logger.Warnf(
				"failed to parse log level '%s', falling back to default level '%s'",
				rootConfig.LogLevel,
				defaultLogLevel,
			)
			logLevel = defaultLogLevel
		}
	}

	log = (&app.LogrusWrapper{Logger: logger}).SetLevel(logLevel)
	log.Infof("Log level is '%s'", logLevel)

	bootstrap.SetLogger(log)

	if len(rootConfig.Servers) == 0 {
		log.Error("no servers are defined in the configuration file")
		os.Exit(1)
	}

	ctx, cancel := context.WithCancel(context.Background())
	wg := &sync.WaitGroup{}

	proxyRoutesInitializer := bootstrap.NewProxyRoutesInitializer(
		bootstrap.DefaultBackendErrorHandler(log.Error),
		bootstrap.DefaultErrorHandler(log.Error),
	)
	staticRoutesInitializer := bootstrap.NewStaticContentRoutesInitializer()

	routesInitializer := bootstrap.HTTPRoutesInitializers{
		proxyRoutesInitializer,
		staticRoutesInitializer,
	}

	for _, serverConfig := range rootConfig.Servers {
		if len(serverConfig.ProxyServices) == 0 && len(serverConfig.StaticServices) == 0 {
			log.Warningf("no services are defined for the server '%s'")
			continue
		}

		router := route.NewDefaultRouter(route.NewWildCardMatcher)

		if err = routesInitializer.InitRoutes(router, serverConfig); err != nil {
			log.Errorf("cannot create server '%s': %s", serverConfig.Name, err)
			os.Exit(1)
		}
		server := &http.Server{
			Addr:         serverConfig.ListenAddress,
			Handler:      router,
			ReadTimeout:  serverConfig.ReadTimeout,
			WriteTimeout: serverConfig.WriteTimeout,
			IdleTimeout:  serverConfig.IdleTimeout,
		}
		wg.Add(1)
		go serve(ctx, serverConfig.Name, wg, server)
	}

	quit := make(chan os.Signal, 1)

	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Info("shutting down")

	cancel()

	wg.Wait()
}

func loadConfig() (*bootstrap.RootConfig, error) {
	const (
		configFileEnv     = "ALTERNEA_CONFIG"
		defaultConfigFile = "alternea.hcl"
	)

	configPath := os.Getenv(configFileEnv)
	if configPath == "" {
		configPath = defaultConfigFile
	}

	fmt.Fprintf(os.Stdout, "using configuration file '%s'\n", configPath)
	return bootstrap.ParseConfig(configPath)

}

func serve(ctx context.Context, name string, wg *sync.WaitGroup, srv *http.Server) {
	var err error
	defer wg.Done()

	go func() {
		if srv.TLSConfig != nil {
			log.Infof("'%s' start listening TLS on %s", name, srv.Addr)
			if err = srv.ListenAndServeTLS("", ""); err != nil && err != http.ErrServerClosed {
				log.Errorf("%s listen: %s", name, err.Error())
				os.Exit(1)
			}
			return
		}
		if err = srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Errorf("%s listen: %s", name, err.Error())
			os.Exit(1)
		}
	}()
	log.Infof("'%s' start listening on %s", name, srv.Addr)

	<-ctx.Done()

	log.Infof("%s server stopped", name)

	ctxShutDown, cancel := context.WithTimeout(context.Background(), shutdownTimeout)

	if err = srv.Shutdown(ctxShutDown); err != nil {
		log.Errorf("%s server shutdown failed: %s", name, err)
	} else {
		log.Infof("%s server exited properly", name)
	}

	defer func() {
		cancel()
	}()
}
