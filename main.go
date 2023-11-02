package db

import (
	"errors"
	"log"
	"os"

	"github.com/go-sql-driver/mysql"
	gorm_mysql "gorm.io/driver/mysql"
	"gorm.io/gorm"
	gorm_logger "gorm.io/gorm/logger"
)

var (
	i *gorm.DB
)

// https://dev.mysql.com/doc/mysql-errors/5.7/en/server-error-reference.html#error_er_bad_db_error
const UnknownDatabaseError = 1049

// Init initialises the database.
func Init(config *Config) error {
	if config.ErrorHandler != nil {
		SetErrorHandler(config.ErrorHandler)
	}

	// https://gorm.io/docs/gorm_config.html
	gormConfig := gorm.Config{
		AllowGlobalUpdate: true,
	}

	gormConfig.Logger = newLogger(
		log.New(os.Stdout, "", 0),
		gorm_logger.Config{
			LogLevel:                  gorm_logger.Warn,
			IgnoreRecordNotFoundError: true, // Checked for explicitly by db.First and co.
			SlowThreshold:             config.SlowThreshold,
			Colorful:                  config.Colour,
		},
	)

	if config.Debug {
		// https://gorm.io/docs/logger.html
		gormConfig.Logger = gormConfig.Logger.LogMode(gorm_logger.Info)
	}

	var err error
	i, err = gorm.Open(gorm_mysql.Open(config.PrimaryDSN()), &gormConfig)

	var needsSeed bool

	if err == nil && config.Fresh {
		if err := dropDatabase(config, &gormConfig); err != nil {
			return err
		}

		if err := createDatabase(config, &gormConfig); err != nil {
			return err
		}

		needsSeed = true
	} else if err != nil {
		// If we fail to connect, it's likely the database doesn't exist yet, so create it.
		// The fallback DSN allows us to connect without the database name.
		var mysqlErr *mysql.MySQLError
		if !errors.As(err, &mysqlErr) || mysqlErr.Number != UnknownDatabaseError {
			return err
		}

		if err := createDatabase(config, &gormConfig); err != nil {
			return err
		}

		i, err = gorm.Open(gorm_mysql.Open(config.PrimaryDSN()), &gormConfig)
		if err != nil {
			return err
		}

		needsSeed = true
	}

	if err := migrate(config); err != nil {
		return err
	}

	if config.Seed != nil && needsSeed {
		return config.Seed()
	}

	return nil
}

func createDatabase(config *Config, gormConfig *gorm.Config) error {
	db, err := gorm.Open(gorm_mysql.Open(config.FallbackDSN()), gormConfig)
	if err != nil {
		return err
	}

	return db.Exec("CREATE DATABASE " + config.Name).Error
}

func dropDatabase(config *Config, gormConfig *gorm.Config) error {
	db, err := gorm.Open(gorm_mysql.Open(config.FallbackDSN()), gormConfig)
	if err != nil {
		return err
	}

	return db.Exec("DROP DATABASE " + config.Name).Error
}

func migrate(config *Config) error {
	return NewMigrator(i, config.Migrations).Migrate(func(db *gorm.DB) error {
		for _, model := range config.Models {
			if err := db.AutoMigrate(model); err != nil {
				return err
			}
		}
		return nil
	})
}
