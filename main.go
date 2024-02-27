package main

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"os"
	"sync"
)


type Service struct {
	Memtable map[string]string
}

func NewService() *Service {
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

	//handler = logMiddleware(handler)
	//handler = authMiddleware(handler)
	return handler
}

// TODO revisit this
func main() {
	println("Hello, World!")

	// TODO LOL this is so bizarre just use SLOG instead
	logger := newLogger(func(s string) {
		println(s)
	})
	service := NewService()


	srv := NewServer(logger, service)
	httpServer := &http.Server{
		Addr:	net.JoinHostPort("localhost", "8080"),
		Handler: srv,
	}

	go func() {
		logger.log("Server is listening on " + httpServer.Addr)
		if err := httpServer.ListenAndServe(); err != nil &&  err != http.ErrServerClosed {
			fmt.Fprintf(os.Stderr, "Could not listen on %s: %v\n", httpServer.Addr, err)
		}
	}()
	
	var wg sync.WaitGroup
	wg.Add(1)

	ctx := context.Background()

	go func(){
		defer wg.Done()

		// TODO What does this do?
		<-ctx.Done()
		logger.log("Shutting down the server...")

		if err := httpServer.Shutdown(ctx); err != nil {
			fmt.Fprintf(os.Stderr, "Error shutting down the server: %v\n", err)
		}
	}()

	wg.Wait()
}
