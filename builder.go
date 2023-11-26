package db

import (
	"errors"

	"gorm.io/gorm"
)

type Map = map[string]any

type Builder[T any] struct {
	query *gorm.DB
}

// B returns a new *Builder prepared for model T.
//
// The builder can be initialised with a where query by passing in a string followed by (optional) args.
//
// Examples:
//
//	db.B[Model]()
//	db.B[Model]("id = ?", 1)
//	db.B[Model]("id = ? and name = ?", 1, "John")
func B[T any](args ...any) *Builder[T] {
	query := i.Model(new(T))

	if len(args) > 0 {
		if s, ok := args[0].(string); ok {
			query = query.Where(s, args[1:]...)
		} else {
			panic("invalid parameter")
		}
	}

	return &Builder[T]{
		query: query,
	}
}

// —————————————————————————————————————————————————————————————————————————————————————————————————
// Chainables
// —————————————————————————————————————————————————————————————————————————————————————————————————

// Debug ensures queries from this builder always log to the terminal.
func (self *Builder[T]) Debug() *Builder[T] {
	return &Builder[T]{
		query: self.query.Debug(),
	}
}

// Unscoped ensures queries include soft-deleted rows. This method does not modify the current
// builder.
func (self *Builder[T]) Unscoped() *Builder[T] {
	return &Builder[T]{
		query: self.query.Unscoped(),
	}
}

// Where adds a WHERE clause to the query.
//
// Examples:
//
//	Where("id = ?", 1)
//	Where("id = ? and name = ?", 1, "John")
func (self *Builder[T]) Where(query string, args ...any) *Builder[T] {
	self.query = self.query.Where(query, args...)
	return self
}

// Order adds an ORDER BY clause to the query.
//
// Examples:
//
//	Order("id ASC")
//	Order("id DESC")
func (self *Builder[T]) Order(value any) *Builder[T] {
	self.query = self.query.Order(value)
	return self
}

// Limit adds a LIMIT clause to the query.
func (self *Builder[T]) Limit(value int) *Builder[T] {
	self.query = self.query.Limit(value)
	return self
}

// Offset adds an OFFSET clause to the query.
func (self *Builder[T]) Offset(offset int) *Builder[T] {
	self.query = self.query.Offset(offset)
	return self
}

// Distinct adds an DISTINCT clause to the query.
func (self *Builder[T]) Distinct() *Builder[T] {
	self.query = self.query.Distinct()
	return self
}

// —————————————————————————————————————————————————————————————————————————————————————————————————
// Finishers
// —————————————————————————————————————————————————————————————————————————————————————————————————

// Find returns all rows that match the query.
func (self *Builder[T]) Find() []*T {
	var rows []*T
	must0(self.query.Find(&rows).Error)
	return rows
}

// First returns the first row that matches the query, and true if it was able to find a row.
func (self *Builder[T]) First() (*T, bool) {
	var row *T
	res := self.query.First(&row)
	if errors.Is(res.Error, gorm.ErrRecordNotFound) {
		return row, false
	}
	must0(res.Error)
	return row, true
}

// Last returns the last row that matches the query, and true if it was able to find a row.
func (self *Builder[T]) Last() (*T, bool) {
	var row *T
	res := self.query.Last(&row)
	if errors.Is(res.Error, gorm.ErrRecordNotFound) {
		return row, false
	}
	must0(res.Error)
	return row, true
}

// Count returns the number of rows that match the query.
func (self *Builder[T]) Count() int64 {
	var n int64
	must0(self.query.Count(&n).Error)
	return n
}

// Exists returns whether any rows exist that match the query.
func (self *Builder[T]) Exists() bool {
	return self.Count() > 0
}

// Update updates the value(s) of column(s) for the rows that match the query. The values
// parameter can take many forms.
//
// Sequence of key/value pairs (like in slog):
//
//	Update("foo", "bar", "n", 123)
//
// Slice of sequence of key/value pairs:
//
//	updates := []any{"foo", "bar", "n", 123}
//	Update(updates...)
//	Update(updates)
//
// Map of map[string]any (aliased to db.Map):
//
//	Update(db.Map{"foo": "bar", "n": 123})
//
// Model (only updates non-zero fields):
//
//	Update(m.Model{Foo: "bar", N: 123})
func (self *Builder[T]) Update(values ...any) {
	f := func(values []any) {
		if (len(values) % 2) != 0 {
			panic("values argument must be an even number of elements (or 1)")
		}

		m := map[string]any{}
		for i := 0; i < len(values); i += 2 {
			m[values[i].(string)] = values[i+1]
		}

		must0(self.query.Updates(m).Error)
	}

	switch len(values) {
	case 0:
		return
	case 1:
		values := values[0]

		if slice, ok := values.([]any); ok {
			f(slice)
		} else {
			must0(self.query.Updates(values).Error)
		}
	default:
		f(values)
	}
}

// Delete soft-deletes all rows that match the query.
func (self *Builder[T]) Delete() {
	must0(self.query.Delete(new(T)).Error)
}

// HardDelete hard-deletes all rows that match the query.
func (self *Builder[T]) HardDelete() {
	must0(self.query.Unscoped().Delete(new(T)).Error)
}
