package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"time"

	. "github.com/dominicfollett/argus-db/service"
)

type Config struct {
	Host     string
	Port     string
	LogLevel slog.Level
	Engine   string
}

var levelMap = map[string]slog.Level{
	"debug": slog.LevelDebug,
	"info":  slog.LevelInfo,
	"warn":  slog.LevelWarn,
	"error": slog.LevelError,
}

func loadConfig(getenv func(string) string) *Config {

	config := &Config{
		Host:     "localhost",
		Port:     "8123",
		LogLevel: slog.LevelInfo,
		Engine:   "naive",
	}

	if host := getenv("HOST"); host != "" {
		config.Host = host
	}

	if port := getenv("PORT"); port != "" {
		config.Port = port
	}

	if logLevel := getenv("LOG_LEVEL"); logLevel != "" {
		config.LogLevel = levelMap[logLevel]
	}

	return config
}

func healthHandler() http.Handler {
	return http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			if r.Method != http.MethodGet {
				w.WriteHeader(http.StatusMethodNotAllowed)
				return
			}

			w.WriteHeader(http.StatusOK)
			w.Write([]byte("OK"))
		},
	)
}

type limitArgs struct {
	Key      string `json:"key"`
	Capacity int    `json:"capacity"`
	Interval int    `json:"interval"`
	Unit     string `json:"unit"`
}

func limitHandler(logger *slog.Logger, service *Service) http.Handler {
	return http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			select {
			case <-r.Context().Done():
				logger.Info("request canceled")
				return
			default:
				if r.Method != http.MethodPost {
					w.WriteHeader(http.StatusMethodNotAllowed)
					return
				}

				// TODO: consider using a pool of buffers with custom decoding, or easyjson or protobuf
				// Ideas: https://github.com/goccy/go-json
				// start := time.Now()
				var args limitArgs
				buffer, err := io.ReadAll(r.Body)
				if err != nil {
					logger.Error("error reading request body", "error", err)
					w.WriteHeader(http.StatusBadRequest)
					w.Write([]byte("error reading request body"))
					return
				}

				if err = json.Unmarshal(buffer, &args); err != nil {
					logger.Error("error decoding request body", "error", err)
					w.WriteHeader(http.StatusBadRequest)
					w.Write([]byte("error decoding request body"))
					return
				}
				// duration := time.Since(start)
				// logger.Info("json.Unmarshal", "duration", duration)

				// Call the service layer
				result, err := service.Limit(r.Context(), args.Key, args.Capacity, args.Interval, args.Unit)
				if err != nil {
					if err.Error() == "request canceled" {
						logger.Info("request canceled")
						return
					} else {
						logger.Error("error calling limiter service: %v", err)
						w.WriteHeader(http.StatusInternalServerError)
						return
					}
				}

				// Write the result
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(result))

			}
		},
	)
}

func loggingMiddleware(logger *slog.Logger, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		next.ServeHTTP(w, r)

		// duration is in nanoseconds
		duration := time.Since(start)
		logger.Info("request", "method", r.Method, "url", r.URL.String(), "duration", duration)
	})
}

func NewServer(logger *slog.Logger, service *Service) http.Handler {
	mux := http.NewServeMux()

	mux.Handle("/api/v1/health", loggingMiddleware(logger, healthHandler()))
	mux.Handle("/api/v1/limit", loggingMiddleware(logger, limitHandler(logger, service)))

	return mux
}

func run(ctx context.Context, getenv func(string) string, stdout io.Writer) error {
	// Create a copy of the parent context that is marked done (its Done channel is closed) when
	// the os.Interrupt signal arrives. This prevents the program from immediately exiting when the os.Interrupt is received
	// which would prevent us from shutting down the server gracefully. "The stop function [cancel] unregisters the signal
	// behavior, which restores the default behavior for a given signal. For example, the default
	// behavior of a Go program receiving os.Interrupt is to exit. Calling NotifyContext(parent, os.Interrupt) will change
	// the behavior to cancel the returned context. Future interrupts received will not trigger the default (exit) behavior
	// until the returned stop function is called." In other words, the stop function cancels the context and restores the
	// default behavior of os.Interrupt on the parent context.
	ctx, cancel := signal.NotifyContext(ctx, os.Interrupt)
	defer cancel()

	config := loadConfig(getenv)

	logger := slog.New(slog.NewJSONHandler(stdout, &slog.HandlerOptions{Level: config.LogLevel}))
	service := NewLimiterService(config.Engine)
	server := NewServer(logger, service)

	// Take note of the timeouts: this makes the server more robust and less susceptible to attacks
	httpServer := &http.Server{
		Addr:              net.JoinHostPort(config.Host, config.Port),
		Handler:           http.TimeoutHandler(server, 1*time.Second, "timeout\n"),
		ReadTimeout:       500 * time.Millisecond,
		ReadHeaderTimeout: 500 * time.Millisecond,
		IdleTimeout:       1 * time.Second,
	}

	// ListenAndServe is a blocking call so we need to run it in a goroutine
	go func() {
		logger.Info("Server is listening on " + httpServer.Addr)
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("Could not listen on %s: %v", httpServer.Addr, err)
		}
	}()

	var wg sync.WaitGroup
	wg.Add(1)

	go func() {
		defer wg.Done()
		// <-ctx.Done() is used within a goroutine to wait for a signal indicating that a context
		// (ctx) has been canceled or has expired. The ctx.Done() method returns a channel that
		// is closed when the context is canceled or when its deadline expires, signaling to the
		// goroutine that it should stop its work and exit.
		// This is a blocking call
		// <- == "listen on" and ctx.Done() == "the channel to listen on", i.e. block until the channel is closed
		<-ctx.Done()

		// Make a new context for the Shutdown because the parent context has already been canceled
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)

		// Don't forget to call cancel to release resources associated with the shutdownCtx
		defer cancel()

		if err := httpServer.Shutdown(shutdownCtx); err != nil {
			slog.Error("error shutting down http server: %v\n", err)
		}
	}()
	wg.Wait()

	return nil
}

func main() {
	ctx := context.Background()
	if err := run(ctx, os.Getenv, os.Stdout); err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
		os.Exit(1)
	}
}
