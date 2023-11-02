package db

// Based on https://github.com/go-gormigrate/gormigrate

import (
	"errors"
	"fmt"
	"reflect"
	"time"

	"gorm.io/gorm"
)

// The marker that indicates that the database schema has been initialised.
const initMigrationID = "SCHEMA_INIT"

var (
	ErrMissingAutoMigrateFunc = errors.New("missing autoMigrate func")
	ErrMissingMigrationID     = func(i int) error { return fmt.Errorf("missing migration ID at index: %d", i) }
	ErrMissingMigrationType   = func(id string) error { return fmt.Errorf("missing migration type for ID: %s", id) }
	ErrReservedMigration      = func(id string) error { return fmt.Errorf("reserved migration ID: %s", id) }
	ErrDuplicateMigration     = func(id string) error { return fmt.Errorf("duplicate migration ID: %s", id) }
)

type MigrateFunc func(*gorm.DB) error

type MigrationType int

const (
	MigrationTypePre MigrationType = iota + 1
	MigrationTypePost
)

type Migration struct {
	// ID of the migration. Must be unique.
	ID string

	// When to run the migration. "Pre" migrations run before the database schema is updated
	// (automigrated) to match any changes to the model structs, and "Post" migrations run
	// afterwards. The decision as to which one to use depends on the purpose of the migration.
	//
	// If a struct field is renamed then the next time the database is automigrated a duplicate
	// column will be added. A pre-migration to rename the column will prevent this.
	//
	// If a struct field is added but needs to be set to some default value then a post-migration
	// would be appropriate.
	Type MigrationType

	// The function to run to migrate the database.
	Run MigrateFunc
}

type Migrator struct {
	db         *gorm.DB
	migrations []*Migration
	tableName  string
	columnName string
	columnSize int
}

// NewMigrator returns a new instance of Migrator.
func NewMigrator(db *gorm.DB, migrations []*Migration) *Migrator {
	return &Migrator{
		db:         db,
		migrations: migrations,
		tableName:  "migrations",
		columnName: "id",
		columnSize: 255,
	}
}

// Migrate migrates the database schema using autoMigrate to run the automigrations.
func (self *Migrator) Migrate(autoMigrate MigrateFunc) error {
	if autoMigrate == nil {
		return ErrMissingAutoMigrateFunc
	}

	err := chain(
		self.checkForInvalidMigrations,
		self.createTable,
	)

	if err != nil {
		return err
	}

	if shouldAutoMigrateOnly, err := self.shouldAutoMigrateOnly(); err != nil {
		return err
	} else if shouldAutoMigrateOnly {
		return self.autoMigrate(autoMigrate)
	}

	return chain(
		func() error { return self.runMigrationType(MigrationTypePre) },
		func() error { return autoMigrate(self.db) },
		func() error { return self.runMigrationType(MigrationTypePost) },
	)
}

func (self *Migrator) runMigrationType(kind MigrationType) error {
	for _, migration := range self.migrations {
		if migration.Type == kind {
			if err := self.runMigration(migration); err != nil {
				return err
			}
		}
	}
	return nil
}

func (self *Migrator) checkForInvalidMigrations() error {
	return chain(
		self.checkForMissingID,
		self.checkForReservedID,
		self.checkForDuplicateID,
		self.checkForMissingType,
	)
}

func (self *Migrator) checkForReservedID() error {
	for _, migration := range self.migrations {
		if migration.ID == initMigrationID {
			return ErrReservedMigration(migration.ID)
		}
	}
	return nil
}

func (self *Migrator) checkForDuplicateID() error {
	lookup := map[string]bool{}

	for _, migration := range self.migrations {
		if lookup[migration.ID] {
			return ErrDuplicateMigration(migration.ID)
		}
		lookup[migration.ID] = true
	}

	return nil
}

func (self *Migrator) checkForMissingType() error {
	for _, migration := range self.migrations {
		if migration.Type == 0 {
			return ErrMissingMigrationType(migration.ID)
		}
	}

	return nil
}

func (self *Migrator) checkForMissingID() error {
	for i, migration := range self.migrations {
		if len(migration.ID) == 0 {
			return ErrMissingMigrationID(i)
		}
	}

	return nil
}

func (self *Migrator) autoMigrate(f MigrateFunc) error {
	return chain(
		func() error { return f(self.db) },
		func() error { return self.insertMigration(initMigrationID, true) },
		func() error {
			for _, migration := range self.migrations {
				if err := self.insertMigration(migration.ID, false); err != nil {
					return err
				}
			}
			return nil
		},
	)
}

func (self *Migrator) runMigration(migration *Migration) error {
	done, err := self.migrationAlreadyRan(migration)
	if err != nil {
		return err
	}

	if !done {
		return chain(
			func() error { return migration.Run(self.db) },
			func() error { return self.insertMigration(migration.ID, true) },
		)
	}

	return nil
}

func (self *Migrator) model() any {
	id := reflect.StructField{
		Name: reflect.ValueOf("ID").Interface().(string),
		Type: reflect.TypeOf(""),
		Tag: reflect.StructTag(fmt.Sprintf(
			`gorm:"primaryKey;column:%s;size:%d"`,
			self.columnName,
			self.columnSize,
		)),
	}

	ts := reflect.StructField{
		Name: reflect.ValueOf("Ts").Interface().(string),
		Type: reflect.TypeOf(time.Time{}),
		Tag:  reflect.StructTag(`gorm:"column:ts"`),
	}

	structType := reflect.StructOf([]reflect.StructField{id, ts})
	structValue := reflect.New(structType).Elem()
	return structValue.Addr().Interface()
}

func (self *Migrator) createTable() error {
	return self.db.Table(self.tableName).AutoMigrate(self.model())
}

func (self *Migrator) migrationAlreadyRan(migration *Migration) (bool, error) {
	var count int64
	err := self.db.
		Table(self.tableName).
		Where(fmt.Sprintf("%s = ?", self.columnName), migration.ID).
		Count(&count).
		Error

	return count > 0, err
}

func (self *Migrator) shouldAutoMigrateOnly() (bool, error) {
	autoMigrated, err := self.migrationAlreadyRan(&Migration{ID: initMigrationID})

	if err != nil {
		return false, err
	}

	if autoMigrated {
		return false, nil
	}

	var count int64
	err = self.db.Table(self.tableName).Count(&count).Error

	return count == 0, err
}

func (self *Migrator) insertMigration(id string, wasRun bool) error {
	row := map[string]any{
		self.columnName: id,
	}

	if wasRun {
		row["ts"] = time.Now()
	}

	return self.db.Table(self.tableName).Create(row).Error
}
