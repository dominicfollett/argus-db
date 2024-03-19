package service

import (
	"context"
	"fmt"
	"time"

	"github.com/dominicfollett/argus-db/database"
)

// Shared Data structure stores the Token Bucket particulars
type Data struct {
	availableTokens int64
	lastRefilled    time.Time // Should this rather be a unix timestamp as int64?
}

type Params struct {
	capacity int64
	interval int32
	unit     string
}

type Service struct {
	Database database.Database
}

func min(a int64, b int64) int64 {
	if a < b {
		return a
	}
	return b
}

// callback is the function that is passed to the database layer
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

	switch p.unit {
	case "s":
		refillTokens = elapsedTime.Seconds() * refillRate
	case "ms":
		refillTokens = float64(elapsedTime.Milliseconds()) * refillRate
	case "us":
		refillTokens = float64(elapsedTime.Microseconds()) * refillRate
	}

	if refillTokens > 0 {
		d.lastRefilled = time.Now()
		d.availableTokens = min(p.capacity, d.availableTokens+int64(refillTokens))
	}

	allowed := d.availableTokens > 0
	if allowed {
		d.availableTokens--
	}

	return d, allowed, nil
}

func NewLimiterService(engine string) *Service {

	return &Service{
		Database: database.NewDatabase(engine, callback),
	}
}

func (s *Service) Limit(ctx context.Context, key string, capacity int64, interval int32, unit string) (string, error) {
	select {
	case <-ctx.Done():
		return "", fmt.Errorf("request canceled")
	default:
		result, err := s.Database.Calculate(key, &Params{capacity, interval, unit})

		if err != nil {
			// TODO
			return "", err
		}

		if result.(bool) {
			return "OK", nil
		} else {
			return "LIMITED", nil
		}
	}
}
