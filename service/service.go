package service

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/dominicfollett/argus-db/database"
)

// Shared Data structure stores the Token Bucket particulars
type Data struct {
	availableTokens int64
	lastRefilled    time.Time // Should this rather be a unix timestamp as int64?
	expiresAt       time.Time
}

type Params struct {
	capacity int64
	interval int32
	unit     string
}

type Service struct {
	database database.Database
	logger   *slog.Logger
}

func min(a int64, b int64) int64 {
	if a < b {
		return a
	}
	return b
}

// evict is a function passed to the database layer
// that determines when a node should be evicted based
// on the data stored at that node.
func evict(data any) bool {
	d := data.(*Data)

	delta := time.Since(d.expiresAt)

	return delta >= 0
}

// callback is the function that is passed to the database layer
// which is invoked on each insert to the DB
func callback(data any, params any) (any, any, error) {
	p := params.(*Params)

	var d *Data
	if data == nil {
		d = &Data{
			availableTokens: p.capacity,
			lastRefilled:    time.Now(),
		}
	} else {
		d = data.(*Data)
	}

	refillRate := float64(p.capacity) / float64(p.interval)
	elapsedTime := time.Since(d.lastRefilled)

	var refillTokens float64
	var unit time.Duration

	switch p.unit {
	case "s":
		refillTokens = elapsedTime.Seconds() * refillRate
		unit = time.Second
	case "ms":
		refillTokens = float64(elapsedTime.Milliseconds()) * refillRate
		unit = time.Millisecond
	case "us":
		refillTokens = float64(elapsedTime.Microseconds()) * refillRate
		unit = time.Microsecond
	}

	// TODO: ideally we should cast this one time only
	if int64(refillTokens) > 0 {
		d.lastRefilled = time.Now()
		d.availableTokens = min(p.capacity, d.availableTokens+int64(refillTokens))
	}

	allowed := d.availableTokens > 0
	if allowed {
		d.availableTokens--
	}

	// If its 1000 tokens every 60 seconds
	// refillRate = 1000 / 60 == 16.666.. tokens/s
	// replenish / refillRate == duration needed to replenish
	replenish := p.capacity - d.availableTokens
	duration := time.Duration(float64(replenish)/refillRate) * unit

	// Set the record's expiry time
	d.expiresAt = time.Now().Add(duration)

	return d, allowed, nil
}

func (s *Service) Shutdown() {
	s.database.Shutdown()
}

func NewLimiterService(engine string, logger *slog.Logger) *Service {

	return &Service{
		database: database.NewDatabase(engine, callback, evict, logger),
		logger:   logger,
	}
}

func (s *Service) Limit(ctx context.Context, key string, capacity int64, interval int32, unit string) (string, error) {
	select {
	case <-ctx.Done():
		return "", fmt.Errorf("request canceled")
	default:
		result, err := s.database.Calculate(key, &Params{capacity, interval, unit})

		if err != nil {
			s.logger.Error("could not calculate rate limit", "error", err)
			return "UNDETERMINED", err
		}

		if result.(bool) {
			return "OK", nil
		} else {
			return "LIMITED", nil
		}
	}
}
