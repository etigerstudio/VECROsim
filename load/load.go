package main

import (
	"bytes"
	"context"
	"errors"
	"io/ioutil"
	"net/http"
	"os"
	"time"
)

func Simulate(ctx context.Context, urlList []string, body []byte, users int, delay time.Duration) {
	for _, url := range urlList{
		for i := 0; i < users; i++ {
			// TODO: configurable request methods
			go singleUser(ctx, "POST", body, url, delay, i)
		}

		// Sleep a little while to avoid congestion
		time.Sleep(time.Duration(int(delay) / (len(url) + 1)))
	}

	select {
	case <-ctx.Done():
		if errors.Is(ctx.Err(), context.DeadlineExceeded) {
			logger.Print("Load simulation completed successfully.")
		}
	}
}

func singleUser(ctx context.Context, method string, body []byte, url string, delay time.Duration, id int) {
	// Perform one request immediately
	performRequest(ctx, method, body, url, id)

	// Perform requests after specified delay afterwards
	t := time.NewTicker(delay)
	for _ = range t.C {
		performRequest(ctx, method, body, url, id)
	}
}

func performRequest(ctx context.Context, method string, body []byte, url string, id int) {
	// Build request with context
	req, err := http.NewRequestWithContext(ctx, method, url, bytes.NewReader(body))
	if err != nil {
		panic(err)
	}

	resp, err := client.Do(req)
	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) {
			// Request has been cancelled
			logger.Printf("[%d] Cancelled", id)
			return
		} else if os.IsTimeout(err) {
			// Request timed out
			logger.Printf("[%d] Timeout", id)
		} else {
			panic(err)
		}
	} else {
		// print "." when requested successfully
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			panic(err)
		}
		logger.Printf("[%d]: %s", id, string(body))
	}
}