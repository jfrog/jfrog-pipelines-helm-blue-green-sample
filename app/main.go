package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
)

const (
	APP_PORT    = "APP_PORT"
	APP_VERSION = "APP_VERSION"
)

var defaultConfig = map[string]string{
	APP_PORT:    "8080",
	APP_VERSION: "0.0.0",
}

var server *http.Server
var stop chan os.Signal

func init() {
	stop = make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
}

func getConfig(key string) string {
	value := os.Getenv(key)
	if value == "" {
		value = defaultConfig[key]
	}
	return value
}

type HttpHandler struct{}

func (h *HttpHandler) ServeHTTP(res http.ResponseWriter, r *http.Request) {
	log.Printf("Request Received: %+v", r)

	switch r.URL.Path {
	case "/ping":
		res.Write([]byte("pong"))
	case "/version":
		res.Write([]byte(getConfig(APP_VERSION)))
	default:
		res.WriteHeader(http.StatusNotFound)
	}
}

func main() {
	port := getConfig(APP_PORT)

	server = &http.Server{Addr: fmt.Sprintf(":%s", port), Handler: &HttpHandler{}}
	go func() {
		log.Printf("Starting http server on port %s", port)
		if err := server.ListenAndServe(); err != nil {
			log.Printf(fmt.Sprintf("Http server stopped: %s", err))
			forceShutdown()
		}
	}()

	waitForShutdown()
}

func waitForShutdown() {
	<-stop
	log.Printf("Interrupt signal received!")
	handleGracefulShutdown()
}

func forceShutdown() {
	stop <- os.Interrupt
}

func handleGracefulShutdown() {
	log.Printf("Shutting down http server...")
	server.Shutdown(context.Background())
}
