package main

import (
	"net/http"
	"os"
	"os/signal"
	"syscall"

	_ "net/http/pprof"

	log "github.com/Sirupsen/logrus"
	"github.com/prometheus/client_golang/prometheus"
)

func main() {
	listen := os.Getenv("VANITY_LISTEN")
	if listen == "" {
		listen = ":8080"
	}

	listenMgmt := os.Getenv("VANITY_LISTEN_MGMT")
	if listenMgmt == "" {
		listenMgmt = ":8081"
	}

	config := loadConfiguration("conf/vanity.yaml")
	srv := NewServer(config.Hosts)

	go handleSigs()
	go func() {
		http.Handle("/metrics", prometheus.Handler())
		http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
			http.Error(w, "OK", http.StatusOK)
		})
		http.HandleFunc("/", http.NotFound)
		log.Fatal(http.ListenAndServe(listenMgmt, nil))
	}()

	s := &http.Server{
		Addr:    listen,
		Handler: prometheus.InstrumentHandler("vanity", srv),
	}
	log.Fatal(s.ListenAndServe())
}

func handleSigs() {
	exitChan := make(chan os.Signal, 10)
	signal.Notify(exitChan, syscall.SIGINT, syscall.SIGTERM, syscall.SIGKILL)

	<-exitChan
	log.Printf("Exiting due to signal")
	os.Exit(0)
}
