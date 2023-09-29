package db

import (
	"gorm.io/gorm"
)

// Instance returns the internal instance of *gorm.DB.
func Instance() *gorm.DB {
	return i
}

// Save updates an existing model, or inserts it if it doesn't already exist.
func Save[T any](model *T) *T {
	must0(i.Save(model).Error)
	return model
}

// FirstOrInit returns the first row that matches the query, or a new prepared instance of T, as
// well as true if a row was found.
func FirstOrInit[T any](value T) (T, bool) {
	var row T
	res := i.Where(value).FirstOrInit(&row)
	must0(res.Error)
	return row, res.RowsAffected > 0
}

// FirstOrCreate returns the first row that matches the query, or creates and returns a new instance
// of T, as well as true if a row was found.
func FirstOrCreate[T any](value T) (T, bool) {
	var row T
	res := i.Where(value).FirstOrCreate(&row)
	must0(res.Error)
	return row, res.RowsAffected == 0
}

// Create creates a new model.
func Create[T any](value *T) *T {
	res := i.Create(&value)
	must0(res.Error)
	return value
}

// CreateInBatches creates multiple new models in batches of batchSize.
func CreateInBatches[T any](values []*T, batchSize int) []*T {
	res := i.CreateInBatches(values, batchSize)
	must0(res.Error)
	return values
}

// Exec executes some raw SQL and returns the number of rows affected.
func Exec(sql string, args ...any) int64 {
	res := i.Exec(sql, args...)
	must0(res.Error)
	return res.RowsAffected
}

// Query executes some raw SQL and returns a scan of the result into T.
func Query[T any](sql string, args ...any) T {
	var value T
	res := i.Raw(sql, args...)
	must0(res.Error)
	must0(res.Scan(&value).Error)
	return value
}
