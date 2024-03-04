package main

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"time"
)

type Service struct {
	Memtable map[string]string
}

func NewLimiterService() *Service {
	return &Service{
		Memtable: make(map[string]string),
	}
}

func (s *Service) Limiter(key string, refill string, capacity string) (string, error) {
	// TODO implement the limiter
	return "OK", nil
}

type Logger struct {
	log func(string)
}

// WTF is this? Use slog instead
func newLogger(logFunc func(string)) *Logger {
	return &Logger{
		log: logFunc,
	}
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}

func NewLimiterHandler(logger *Logger, service *Service) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		// Get the query parameters
		query := r.URL.Query()

		// TODO add more parameters
		key := query.Get("key")
		refill := query.Get("refill")
		capacity := query.Get("capacity")

		if key == "" {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("key is required"))
			return
		}

		if refill == "" {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("refill is required"))
			return
		}

		if capacity == "" {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("capacity is required"))
			return
		}

		// Call the service layer
		result, err := service.Limiter(key, refill, capacity)

		// TODO WTF is this?
		logger.log("An internal error occurred.")

		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
			return
		}

		// Write the result
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(result))
	}

}

func addRoutes(mux *http.ServeMux, logger *Logger, service *Service) {
	mux.HandleFunc("/api/v1/health", healthHandler)
	mux.HandleFunc("/api/v1/limit", NewLimiterHandler(logger, service))
}

func NewServer(logger *Logger, service *Service) http.Handler {
	mux := http.NewServeMux()

	addRoutes(mux, logger, service)

	var handler http.Handler = mux

	handler = loggerMiddleware(handler)
	//handler = authMiddleware(handler)
	return handler
}

func run(
	ctx context.Context,
	args []string,
	getenv func(string) string,
	stdout io.Writer,
	stderr io.Writer,
) error {
	ctx, cancel := signal.NotifyContext(ctx, os.Interrupt)
	defer cancel()

	// args override environment variables?
	// Maybe just use simple args for now
	// I should probably define a struct for the configuration?
	// And then initialize it from the environment variables and the command line arguments
	// options could include the port, the log level, the database type and connection string, etc.
	// a path to a file containing the database configuration could also be an option

	// TODO: Get the logger level from the configuration
	loggerLevel := slog.LevelInfo
	logger := slog.New(slog.NewJSONHandler(stdout, &slog.HandlerOptions{Level: loggerLevel}))

	service := NewLimiterService()

	srv := NewServer(logger, service)
	httpServer := &http.Server{
		Addr:    net.JoinHostPort("localhost", "8080"),
		Handler: srv,
	}

	go func() {
		logger.Info("Server is listening on " + httpServer.Addr)
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			fmt.Fprintf(stderr, "Could not listen on %s: %v\n", httpServer.Addr, err)
		}
	}()

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		<-ctx.Done()
		// make a new context for the Shutdown (thanks Alessandro Rosetti)
		shutdownCtx := context.Background()
		shutdownCtx, cancel := context.WithTimeout(shutdownCtx, 10*time.Second)
		defer cancel()
		if err := httpServer.Shutdown(shutdownCtx); err != nil {
			fmt.Fprintf(stderr, "error shutting down http server: %s\n", err)
		}
	}()
	wg.Wait()
	return nil
}

func main() {
	ctx := context.Background()
	if err := run(ctx, os.Args, os.Getenv, os.Stdout, os.Stderr); err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
		os.Exit(1)
	}
}
