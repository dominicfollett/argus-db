package database

import "github.com/dominicfollett/argus-db/database/naive"

type Database interface {
	Insert(key string, capacity int, interval int, unit string) (bool, error)
}

func NewDatabase(engine string) Database {
	switch engine {
	case "naive":
		return naive.NewDB()
	default:
		return naive.NewDB()
	}
}
