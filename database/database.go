package database

import (
	"log/slog"

	"github.com/dominicfollett/argus-db/database/naive"
)

type Database interface {
	Calculate(key string, params any) (any, error)
	Shutdown()
}

func NewDatabase(
	engine string,
	callback func(data any, params any) (any, any, error),
	logger *slog.Logger,
) Database {

	switch engine {
	case "naive":
		return naive.NewDB(callback, logger)
	default:
		return naive.NewDB(callback, logger)
	}
}
