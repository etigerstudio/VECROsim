package main

import (
	"context"
	"crypto/tls"
	"flag"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"
)

var tr http.RoundTripper = &http.Transport{
	TLSClientConfig:   &tls.Config{InsecureSkipVerify: true},
	DisableKeepAlives: true,
}
var client = &http.Client{
	Transport: tr,
	Timeout: 5 * time.Second,
}
var logger = log.New(os.Stderr, "", 0)

func main() {
	usersPtr := flag.Int("users", 1, "Number of concurrent users")
	delayPtr := flag.Duration("delay", time.Second, "Delay between calls per user (ms)")
	urlPtr := flag.String("url", "http://127.0.0.1", "URL to perform requests on")
	bodyPtr := flag.String("body", "", "Request body")
	durationPtr := flag.Duration("duration", 0, "Duration of this load simulation")

	flag.Parse()

	// Make Ctrl-C interruptible
	ctx := interruptibleCxt()
	// Cancel requests when duration expired
	if *durationPtr != 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, *durationPtr)
		defer cancel()
	}

	Simulate(ctx, *urlPtr, []byte(*bodyPtr), *usersPtr, *delayPtr)
}

// TODO: Migrate to main
func interruptibleCxt() context.Context {
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt, os.Kill)

	ctx, cancel := context.WithCancel(context.Background())

	go func() {
		defer cancel()
		_ = <-sig
		logger.Println("Load simulation cancelled.")
	}()

	return ctx
}