package main

import (
	"fmt"
	"net"
	"net/http"
	"sync"
	"time"

	"golang.org/x/time/rate"
)

func (app *application) recoverPanic(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Create a deferred function (which will always be run in the event of a panic
		// as Go unwinds the stack).
		defer func() {
			// Use the builtin recover function to check if there has been a panic or
			// not.
			if err := recover(); err != nil {
				// If there was a panic, set a "Connection: close" header on the
				// response. This acts as a trigger to make Go's HTTP server
				// automatically close the current connection after a response has been
				// sent.
				w.Header().Set("Connection", "close")
				// The value returned by recover() has the type interface{}, so we use
				// fmt.Errorf() to normalize it into an error and call our
				// serverErrorResponse() helper. In turn, this will log the error using
				// our custom Logger type at the ERROR level and send the client a 500
				// Internal Server Error response.
				app.serverErrorResponse(w, r, fmt.Errorf("%s", err))
			}
		}()
		next.ServeHTTP(w, r)
	})
}

func (app *application) rateLimit(next http.Handler) http.Handler {

	type client struct {
		limiter  *rate.Limiter
		lastSeen time.Time
	}

	// concurrent map for storing clients and their rate limiter
	var clients sync.Map

	// go routine to reduce map size in runtime
	go func() {
		for {
			time.Sleep(time.Minute)
			// remove old clients from map
			clients.Range(func(key, value interface{}) bool {
				ip := key.(string)
				client := value.(*client)
				if time.Since(client.lastSeen) > 3*time.Minute {
					clients.Delete(ip)
				}
				return true
			})
		}
	}()

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		ip, _, err := net.SplitHostPort(r.RemoteAddr)
		if err != nil {
			app.serverErrorResponse(w, r, err)
			return
		}

		_, ok := clients.Load(ip)
		if !ok {
			limiter := rate.NewLimiter(2, 4)
			clients.Store(ip, &client{limiter: limiter, lastSeen: time.Now()})
		}

		value, ok := clients.Load(ip)
		if !ok {
			app.serverErrorResponse(w, r, err)
			return
		}

		existingClient := value.(*client)
		existingClient.lastSeen = time.Now()

		if !existingClient.limiter.Allow() {
			app.rateLimitExceededResponse(w, r)
			return
		}

		next.ServeHTTP(w, r)
	})
}