package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"slices"
	"strings"
	"time"

	"github.com/felixge/httpsnoop"
	"github.com/henrtytanoh/greenlight/internal/data"
	"github.com/henrtytanoh/greenlight/internal/validator"
	"github.com/tomasen/realip"
)

const (
	keyWithoutTTLVal = -1
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

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		// Implement Fixed window strategy rate limiting with redis
		// Functional, but does not handle bursty traffic and can result in 2* traffic at the frontier on the previous and current window
		if app.config.limiter.enabled {
			ip := realip.FromRequest(r)
			key := "ratelimit:" + ip

			// Using redis pipeline as we need to execute INCR and TTL. Note pipelines are not transactions, we still need to
			// check for errors on both steps
			ctx := context.Background()
			requestsCount, err := app.redisClient.Incr(ctx, key).Result()
			if err != nil {
				app.serverErrorResponse(w, r, err)
				return
			}
			// Set TTL if not present(i.e at the start of a window)
			if ttl, err := app.redisClient.TTL(ctx, key).Result(); err != nil {
				app.serverErrorResponse(w, r, err)
				return
			} else {
				if ttl == keyWithoutTTLVal {
					if err := app.redisClient.Expire(ctx, key, time.Duration(app.config.limiter.windowLength)*time.Second).Err(); err != nil {
						app.serverErrorResponse(w, r, err)
						return
					}
				}
			}

			if requestsCount > int64(app.config.limiter.requestLimit) {
				app.rateLimitExceededResponse(w, r)
				return
			}
		}
		next.ServeHTTP(w, r)
	})
}

func (app *application) authenticate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Add the "Vary: Authorization" header to the response. This indicates to any
		// caches that the response may vary based on the value of the Authorization
		// header in the request
		w.Header().Add("Vary", "Authorization")
		authorizationHeader := r.Header.Get("Authorization")

		if authorizationHeader == "" {
			r = app.contextSetUser(r, data.AnonymousUser)
			next.ServeHTTP(w, r)
			return
		}

		headersParts := strings.Split(authorizationHeader, " ")
		if len(headersParts) != 2 || headersParts[0] != "Bearer" {
			app.invalidAuthenticationTokenResponse(w, r)
			return
		}

		token := headersParts[1]

		v := validator.New()

		if data.ValidateTokenPlaintext(v, token); !v.Valid() {
			app.invalidAuthenticationTokenResponse(w, r)
			return
		}

		user, err := app.models.Users.GetForToken(data.ScopeAuthentication, token)
		if err != nil {
			switch {
			case errors.Is(err, data.ErrRecordNotFound):
				app.invalidAuthenticationTokenResponse(w, r)
			default:
				app.serverErrorResponse(w, r, err)
			}
			return
		}

		r = app.contextSetUser(r, user)
		next.ServeHTTP(w, r)
	})
}

func (app *application) requireAuthenticatedUser(next http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user := app.contextGetUser(r)
		if user.IsAnonymous() {
			app.authenticationRequiredResponse(w, r)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// Checks that a user is both authenticated and activated.
func (app *application) requireActivatedUser(next http.HandlerFunc) http.HandlerFunc {
	// Rather than returning this http.HandlerFunc we assign it to the variable fn.
	fn := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user := app.contextGetUser(r)
		// Check that a user is activated.
		if !user.Activated {
			app.inactiveAccountResponse(w, r)
			return
		}
		next.ServeHTTP(w, r)
	})
	// Wrap fn with the requireAuthenticatedUser() middleware before returning it.
	return app.requireAuthenticatedUser(fn)
}

func (app *application) requirePermission(code string, next http.HandlerFunc) http.HandlerFunc {
	fn := func(w http.ResponseWriter, r *http.Request) {
		user := app.contextGetUser(r)

		permissions, err := app.models.Permissions.GetAllForUser(user.ID)
		if err != nil {
			app.serverErrorResponse(w, r, err)
			return
		}

		if !permissions.Include(code) {
			app.notPermittedResponse(w, r)
			return
		}

		next.ServeHTTP(w, r)
	}

	return app.requireActivatedUser(fn)
}

func (app *application) enableCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Vary", "Origin")

		origin := r.Header.Get("Origin")

		if len(app.config.cors.trustedOrigins) != 0 {
			if slices.Contains(app.config.cors.trustedOrigins, origin) {
				w.Header().Set("Access-Control-Allow-Origin", origin)

				// check if the request has the HTTP method OPTIONS and contains the
				// Access-Control-Request-Headers header
				if r.Method == http.MethodOptions && r.Header.Get("Access-Control-Request-Headers") != "" {
					// Set the appropriate headers to allow the browser to make requests
					w.Header().Set("Access-Control-Allow-Methods", "OPTIONS, PUT, PATCH, DELETE")
					w.Header().Set("Access-Control-Allow-Headers", "Authorization, Content-Type")

					w.WriteHeader(http.StatusOK)
					return
				}
			}
		}
		next.ServeHTTP(w, r)
	})
}

func (app *application) metrics(next http.Handler) http.Handler {

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		app.m.totalRequestReceived.Inc()
		metrics := httpsnoop.CaptureMetrics(next, w, r)

		app.m.totalResponsesSent.Inc()
		// check 500 status code
		if metrics.Code == http.StatusInternalServerError {
			app.m.totalInternalServer.Inc()
		}
		if metrics.Code >= http.StatusBadRequest && metrics.Code < http.StatusInternalServerError {
			app.m.totalClientSideError.Inc()
		}
		app.m.requestDuration.Set(metrics.Duration.Seconds())
	})
}
