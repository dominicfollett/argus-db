package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net"
	"net/http"

	// _ "net/http/pprof".
	"os"
	"os/signal"
	"sync"
	"time"

	"github.com/dominicfollett/argus-db/service"
)

// http.Server Timeouts.
const ReadTimeout = 500 * time.Millisecond
const ReadHeaderTimeout = 500 * time.Millisecond
const IdleTimeout = 2 * time.Second // TODO: tune this.

// Server Shutdown timeout.
const ShutdownTimeout = 10 * time.Second

// Keep it simple - we don't need more than this.
type Config struct {
	Host     string
	Port     string
	LogLevel slog.Level
	Engine   string
}

// Keep it simple.
func loadConfig(getenv func(string) string) *Config {
	levelMap := map[string]slog.Level{
		"debug": slog.LevelDebug,
		"info":  slog.LevelInfo,
		"warn":  slog.LevelWarn,
		"error": slog.LevelError,
	}

	config := &Config{
		Host:     "0.0.0.0",
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

func healthHandler(logger *slog.Logger) http.Handler {
	return http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			if r.Method != http.MethodGet {
				w.WriteHeader(http.StatusMethodNotAllowed)
				return
			}

			w.WriteHeader(http.StatusOK)
			_, err := w.Write([]byte("OK"))
			if err != nil {
				logger.Error("[healthHandler] error writing response", "error", err)
			}
		},
	)
}

type limitArgs struct {
	Key      string `json:"key"`
	Capacity int64  `json:"capacity"`
	Interval int32  `json:"interval"`
	Unit     string `json:"unit"`
}

func limitHandler(logger *slog.Logger, s *service.Service) http.Handler {
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

					_, err = w.Write([]byte("error reading request body"))
					if err != nil {
						logger.Error("[limitHandler] error writing response", "error", err)
					}
					return
				}

				if err = json.Unmarshal(buffer, &args); err != nil {
					logger.Error("error decoding request body", "error", err)
					w.WriteHeader(http.StatusBadRequest)

					_, err = w.Write([]byte("error decoding request body"))
					if err != nil {
						logger.Error("[limitHandler] error writing response", "error", err)
					}
					return
				}
				// duration := time.Since(start)
				// logger.Info("json.Unmarshal", "duration", duration)

				// Call the service layer
				result, err := s.Limit(r.Context(), args.Key, args.Capacity, args.Interval, args.Unit)
				if err != nil {
					if err.Error() == "request canceled" {
						logger.Info("request canceled")
						return
					}

					logger.Error("error calling limiter service: %v", err)
					w.WriteHeader(http.StatusInternalServerError)
					return
				}

				// Write the result
				w.WriteHeader(http.StatusOK)
				_, err = w.Write([]byte(result))
				if err != nil {
					logger.Error("[limitHandler] error writing response", "error", err)
				}
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

func NewServer(logger *slog.Logger, s *service.Service) http.Handler {
	mux := http.NewServeMux()

	mux.Handle("/api/v1/health", loggingMiddleware(logger, healthHandler(logger)))
	mux.Handle("/api/v1/limit", limitHandler(logger, s))

	return mux
}

func run(ctx context.Context, getenv func(string) string, stdout io.Writer) error {
	var err error

	ctx, cancel := signal.NotifyContext(ctx, os.Interrupt)
	defer cancel()

	config := loadConfig(getenv)

	logger := slog.New(slog.NewJSONHandler(stdout, &slog.HandlerOptions{Level: config.LogLevel}))
	s := service.NewLimiterService(config.Engine, logger)
	server := NewServer(logger, s)

	// Take note of the timeouts: this makes the server more robust and less susceptible to attacks
	httpServer := &http.Server{
		Addr:              net.JoinHostPort(config.Host, config.Port),
		Handler:           http.TimeoutHandler(server, 1*time.Second, "timeout\n"),
		ReadTimeout:       ReadTimeout,
		ReadHeaderTimeout: ReadHeaderTimeout,
		IdleTimeout:       IdleTimeout,
	}

	go func() {
		logger.Info("server is listening on " + httpServer.Addr)
		if err = httpServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Error("could not listen on:", "address", httpServer.Addr, "error", err)
		}
	}()

	// Profiling
	// go func() {
	//	http.ListenAndServe(":6060", nil)
	// }()

	var wg sync.WaitGroup
	wg.Add(1)

	go func() {
		defer wg.Done()

		// Wait for cancel signal
		<-ctx.Done()

		// Make a new context for the Shutdown because the parent context has already been canceled
		shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), ShutdownTimeout)

		// Don't forget to call cancel to release resources associated with the shutdownCtx
		defer shutdownCancel()

		logger.Info("shutting down http server")
		if err = httpServer.Shutdown(shutdownCtx); err != nil {
			logger.Error("error shutting down http server", "error", err)
		}

		logger.Info("shutting down rate limiter service")
		s.Shutdown()
	}()
	wg.Wait()

	return err
}

func main() {
	ctx := context.Background()
	if err := run(ctx, os.Getenv, os.Stdout); err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
		os.Exit(1)
	}
}
