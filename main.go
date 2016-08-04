package main

import (
	"net/http"
	"os"
	"os/signal"
	"syscall"

	_ "net/http/pprof"

	"github.com/dominikschulz/vanity/server"
	"github.com/go-kit/kit/log"
	"github.com/prometheus/client_golang/prometheus"
)

func main() {
	logger := log.NewJSONLogger(os.Stdout)

	listen := os.Getenv("VANITY_LISTEN")
	if listen == "" {
		listen = ":8080"
	}

	listenMgmt := os.Getenv("VANITY_LISTEN_MGMT")
	if listenMgmt == "" {
		listenMgmt = ":8081"
	}

	config, err := loadConfiguration(logger, "conf/vanity.yaml")
	if err != nil {
		logger.Log("level", "error", "msg", "Could not load config", "err", err)
		os.Exit(1)
	}

	srv := server.New(server.Config{
		Log:   logger,
		Hosts: config.Hosts,
	})

	go handleSigs()
	go func() {
		http.Handle("/metrics", prometheus.Handler())
		http.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
			http.Error(w, "OK", http.StatusOK)
		})
		http.HandleFunc("/", http.NotFound)
		if err := http.ListenAndServe(listenMgmt, nil); err != nil {
			logger.Log("level", "error", "msg", "Failed to listen on management port", "err", err)
		}
	}()

	s := &http.Server{
		Addr:    listen,
		Handler: prometheus.InstrumentHandler("vanity", srv),
	}
	if err := s.ListenAndServe(); err != nil {
		logger.Log("level", "error", "msg", "Failed to listen", "err", err)
		os.Exit(1)
	}
}

func handleSigs() {
	exitChan := make(chan os.Signal, 10)
	signal.Notify(exitChan, syscall.SIGINT, syscall.SIGTERM, syscall.SIGKILL)

	<-exitChan
	os.Exit(0)
}
