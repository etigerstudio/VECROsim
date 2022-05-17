package main

import (
	"context"
	"crypto/tls"
	"flag"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"time"
)

var tr http.RoundTripper = &http.Transport{
	TLSClientConfig:   &tls.Config{InsecureSkipVerify: true},
	DisableKeepAlives: true,
}
var client = &http.Client{
	Transport: tr,
	Timeout: 15 * time.Second,
}
var logger = log.New(os.Stderr, "", 0)

const urlSeparator  = " "

func main() {
	usersPtr := flag.Int("users", 1, "Number of concurrent users")
	delayPtr := flag.Duration("delay", time.Second, "Delay between calls per user (ms)")
	urlListPtr := flag.String("url", "http://127.0.0.1", "URLs to perform requests on.\nSeparate each URLs by a whitespace if there're multiple URLs to request on.\n")
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

	Simulate(ctx, parseURLList(urlListPtr), []byte(*bodyPtr), *usersPtr, *delayPtr)
}

func parseURLList(str *string) []string {
	if *str == "" {
		panic("empty URL given")
	}

	return strings.Split(*str, urlSeparator)
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