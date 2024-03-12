package service

import (
	"context"
	"fmt"
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

func NewLimiterService(engine string) *Service {

	return &Service{
		Database: NewDatabase(engine),
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
