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

func Simulate(ctx context.Context, url string, body []byte, users int, delay time.Duration) {
	req, err := http.NewRequest("POST", url, bytes.NewReader(body))  // TODO: configurable request methods
	if err != nil {
		panic(err)
	}
	
	req = req.WithContext(ctx)
	for i := 0; i < users; i++ {
		go singleUser(req, delay, i)
	}

	select {
	case <-ctx.Done():
		if errors.Is(ctx.Err(), context.DeadlineExceeded) {
			logger.Print("Load simulation completed successfully.")
		}
	}
}

func singleUser(req *http.Request, delay time.Duration, id int) {
	// perform one request immediately
	performRequest(req, id)

	// perform requests after specified delay afterwards
	t := time.NewTicker(delay)
	for _ = range t.C {
		performRequest(req, id)
	}
}

func performRequest(req *http.Request, id int) {
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
		_, err = ioutil.ReadAll(resp.Body)
		logger.Printf("[%d] .", id)
	}
}