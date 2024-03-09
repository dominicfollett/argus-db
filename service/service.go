package service

import (
	"context"
	"fmt"

	. "github.com/dominicfollett/argus-db/database"
)

type Service struct {
	Database Database
}

func NewLimiterService(engine string) *Service {

	return &Service{
		Database: NewDatabase(engine),
	}
}

func (s *Service) Limit(ctx context.Context, key string, capacity int, interval int, unit string) (string, error) {
	select {
	case <-ctx.Done():
		return "", fmt.Errorf("request canceled")
	default:
		// TODO Validate the input?

		result, err := s.Database.Insert(key, capacity, interval, unit)
		if err != nil {
			// TODO
			return "", err
		}

		if result {
			return "OK", nil
		} else {
			return "LIMITED", nil
		}
	}
}
