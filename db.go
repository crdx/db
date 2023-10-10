package db

import (
	"gorm.io/gorm"
)

// Instance returns the internal instance of *gorm.DB.
func Instance() *gorm.DB {
	return i
}

type ID interface {
	~string | ~uint | ~int
}

// Interface Model represents an instance of a model object. These will normally be implemented
// by calling the Builder method on db.For[T](self.ID), but Delete may also do other work to
// maintain referential integrity.
type Model interface {
	Update(...any)
	Delete()
}

// —————————————————————————————————————————————————————————————————————————————————————————————————
// For
// —————————————————————————————————————————————————————————————————————————————————————————————————

// For[T] returns a *Builder[T] for the T with the specified ID. The ID can be provided as a value
// that satisfies the interface ID (~string, ~uint, ~int).
func For[T any, U ID](id U) *Builder[T] {
	return B[T]().Where("id", must(ToUint(id)))
}

// ForD[T] returns a *Builder[T] (in debug mode) for the T with the specified ID. The ID can be
// provided as a value that satisfies the interface ID (~string, ~uint, ~int)
func ForD[T any, U ID](id U) *Builder[T] {
	return For[T](id).Debug().Where("id", must(ToUint(id)))
}

// —————————————————————————————————————————————————————————————————————————————————————————————————
// First
// —————————————————————————————————————————————————————————————————————————————————————————————————

// First[T] returns *T for the T with the specified ID, and true if it was found. The ID can be
// provided as a value that satisfies the interface ID (~string, ~uint, ~int)
func First[T any, U ID](id U) (*T, bool) {
	if id, err := ToUint(id); err != nil {
		return nil, false
	} else {
		return For[T](id).First()
	}
}

// FirstD[T] returns *T for the T with the specified ID (in debug mode), and true if it was found.
// The ID can be provided as a value that satisfies the interface ID (~string, ~uint, ~int)
func FirstD[T any, U ID](id U) (*T, bool) {
	if id, err := ToUint(id); err != nil {
		return nil, false
	} else {
		return ForD[T](id).First()
	}
}

// —————————————————————————————————————————————————————————————————————————————————————————————————
// Save
// —————————————————————————————————————————————————————————————————————————————————————————————————

func save[T any](i *gorm.DB, model *T) *T {
	must0(i.Save(model).Error)
	return model
}

// Save updates an existing model, or inserts it if it doesn't already exist.
func Save[T any](model *T) *T {
	return save(i, model)
}

// SaveD updates an existing model (in debug mode), or inserts it if it doesn't already exist.
func SaveD[T any](model *T) *T {
	return save(i.Debug(), model)
}

// —————————————————————————————————————————————————————————————————————————————————————————————————
// FirstOrInit
// —————————————————————————————————————————————————————————————————————————————————————————————————

func firstOrInit[T any](i *gorm.DB, value T) (*T, bool) {
	var row *T
	res := i.Where(value).FirstOrInit(&row)
	must0(res.Error)
	return row, res.RowsAffected > 0
}

// FirstOrInit returns the first row that matches the query, or a new prepared instance of T, as
// well as true if a row was found.
func FirstOrInit[T any](value T) (*T, bool) {
	return firstOrInit(i, value)
}

// FirstOrInitD returns the first row that matches the query (in debug mode), or a new prepared
// instance of T, as well as true if a row was found.
func FirstOrInitD[T any](value T) (*T, bool) {
	return firstOrInit(i.Debug(), value)
}

// —————————————————————————————————————————————————————————————————————————————————————————————————
// FirstOrCreate
// —————————————————————————————————————————————————————————————————————————————————————————————————

func firstOrCreate[T any](i *gorm.DB, value T) (*T, bool) {
	var row *T
	res := i.Where(value).FirstOrCreate(&row)
	must0(res.Error)
	return row, res.RowsAffected == 0
}

// FirstOrCreate returns the first row that matches the query, or creates and returns a new instance
// of T, as well as true if a row was found.
func FirstOrCreate[T any](value T) (*T, bool) {
	return firstOrCreate(i, value)
}

// FirstOrCreateD returns the first row that matches the query (in debug mode), or creates and
// returns a new instance of T, as well as true if a row was found.
func FirstOrCreateD[T any](value T) (*T, bool) {
	return firstOrCreate(i.Debug(), value)
}

// —————————————————————————————————————————————————————————————————————————————————————————————————
// Create
// —————————————————————————————————————————————————————————————————————————————————————————————————

func create[T any](i *gorm.DB, value *T) *T {
	res := i.Create(&value)
	must0(res.Error)
	return value
}

// Create creates a new model.
func Create[T any](value *T) *T {
	return create(i, value)
}

// CreateD creates a new model (in debug mode).
func CreateD[T any](value *T) *T {
	return create(i.Debug(), value)
}

// —————————————————————————————————————————————————————————————————————————————————————————————————
// CreateInBatches
// —————————————————————————————————————————————————————————————————————————————————————————————————

func createInBatches[T any](i *gorm.DB, values []*T, batchSize int) []*T {
	res := i.CreateInBatches(values, batchSize)
	must0(res.Error)
	return values
}

// CreateInBatches creates multiple new models in batches of batchSize.
func CreateInBatches[T any](values []*T, batchSize int) []*T {
	return createInBatches(i, values, batchSize)
}

// CreateInBatchesD creates multiple new models (in debug mode) in batches of batchSize.
func CreateInBatchesD[T any](values []*T, batchSize int) []*T {
	return createInBatches(i.Debug(), values, batchSize)
}

// —————————————————————————————————————————————————————————————————————————————————————————————————
// Exec
// —————————————————————————————————————————————————————————————————————————————————————————————————

func exec(i *gorm.DB, sql string, args ...any) int64 {
	res := i.Exec(sql, args...)
	must0(res.Error)
	return res.RowsAffected
}

// Exec executes some raw SQL and returns the number of rows affected.
func Exec(sql string, args ...any) int64 {
	return exec(i, sql, args...)
}

// ExecD executes some raw SQL (in debug mode) and returns the number of rows affected.
func ExecD(sql string, args ...any) int64 {
	return exec(i.Debug(), sql, args...)
}

// —————————————————————————————————————————————————————————————————————————————————————————————————
// Query
// —————————————————————————————————————————————————————————————————————————————————————————————————

func query[T any](i *gorm.DB, sql string, args ...any) T {
	var value T
	res := i.Raw(sql, args...)
	must0(res.Error)
	must0(res.Scan(&value).Error)
	return value
}

// Query executes some raw SQL and returns a scan of the result into T.
func Query[T any](sql string, args ...any) T {
	return query[T](i, sql, args...)
}

// QueryD executes some raw SQL (in debug mode) and returns a scan of the result into T.
func QueryD[T any](sql string, args ...any) T {
	return query[T](i.Debug(), sql, args...)
}
