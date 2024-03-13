package service

import (
	"context"
	"fmt"
	"math"
	"time"

	. "github.com/dominicfollett/argus-db/database"
)

// Shared Data structure stores the Token Bucket particulars
type Data struct {
	capacity  int32
	timestamp time.Time // Should this rather be a unix timestamp as int64?
}

type Params struct {
	capacity int32
	interval int32
	unit     string
}

type Service struct {
	Database Database
}

// TODO: This is a shit show, CLEAN
func callback(data any, params any) (any, any, error) {
	p := params.(*Params)

	var d *Data
	if data == nil {
		d = &Data{
			capacity:  p.capacity,
			timestamp: time.Now(),
		}
	} else {
		d = data.(*Data)
	}

	refill_rate := float64(p.capacity) / float64(p.interval)

	available_tokens := float64(d.capacity)

	last_refilled := d.timestamp

	elapsed_time := time.Since(last_refilled)

	var refill_tokens float64

	switch p.unit {
	case "s":
		refill_tokens = elapsed_time.Seconds() * refill_rate
	case "ms":
		refill_tokens = float64(elapsed_time.Milliseconds()) * refill_rate
	case "us":
		refill_tokens = float64(elapsed_time.Microseconds()) * refill_rate
	}

	if refill_tokens > 0 {
		d.timestamp = time.Now()
		available_tokens = math.Min(float64(p.capacity), available_tokens+refill_tokens)
	}

	newCapacity := int32(available_tokens)

	fmt.Println("New Capacity: ", newCapacity)

	allowed := newCapacity > 0
	if allowed {
		newCapacity--
	}

	d.capacity = newCapacity

	return d, allowed, nil
}

func NewLimiterService(engine string) *Service {

	return &Service{
		Database: NewDatabase(engine, callback),
	}
}

func (s *Service) Limit(ctx context.Context, key string, capacity int32, interval int32, unit string) (string, error) {
	select {
	case <-ctx.Done():
		return "", fmt.Errorf("request canceled")
	default:
		// TODO Validate the input?

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
