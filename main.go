package main

import (
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	listen := os.Getenv("VANITY_LISTEN")
	if listen == "" {
		listen = ":8080"
	}

	config := loadConfiguration("conf/vanity.yaml")

	s := NewServer(config.Hosts)

	go handleSigs()

	log.Fatal(http.ListenAndServe(listen, s))
}

func handleSigs() {
	exitChan := make(chan os.Signal, 10)
	signal.Notify(exitChan, syscall.SIGINT, syscall.SIGTERM, syscall.SIGKILL)

	<-exitChan
	log.Printf("Exiting due to signal")
	os.Exit(0)
}
