package database

import "github.com/dominicfollett/argus-db/database/naive"

type Database interface {
	Calculate(key string, params any) (any, error)
}

func NewDatabase(engine string, callback func(data any, params any) (any, any, error)) Database {
	switch engine {
	case "naive":
		return naive.NewDB(callback)
	default:
		return naive.NewDB(callback)
	}
}
