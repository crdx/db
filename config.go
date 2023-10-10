package db

import (
	"fmt"
	"strings"
	"time"

	"github.com/go-gormigrate/gormigrate/v2"
)

type Config struct {
	Name          string                  // The database name.
	User          string                  // The database username.
	Pass          string                  // The database password.
	Host          string                  // The database hostname.
	Socket        string                  // The database socket path.
	CharSet       string                  // The database character set.
	TimeZone      string                  // The database timezone.
	Models        []Model                 // A list of models to migrate.
	Migrations    []*gormigrate.Migration // A list of manual migrations to run.
	Debug         bool                    // Whether to log queries.
	Colour        bool                    // Whether to display colour in debugging output.
	Fresh         bool                    // Whether to drop and recreate the database (for tests).
	ErrorHandler  func(err error)         // A function to run if a database error occurs.
	SlowThreshold time.Duration           // Threshold for queries to be considered slow.
}

// PrimaryDSN returns the DSN with the database name specified.
func (self *Config) PrimaryDSN() string {
	dsn, _ := self.buildDSNs()
	return dsn
}

// FallbackDSN returns the DSN without the database name specified.
func (self *Config) FallbackDSN() string {
	_, dsn := self.buildDSNs()
	return dsn
}

// buildDSNs builds two DSNs: one with the database name ("primary"), and one without ("fallback).
func (self *Config) buildDSNs() (primaryDSN string, fallbackDSN string) {
	// https://github.com/go-sql-driver/mysql#dsn-data-source-name
	// {DB_USER}:{DB_PASS}@unix({DB_SOCK})/{DB_NAME}?parseTime=True&charset={DB_CHARSET}&loc={DB_TZ}
	// {DB_USER}:{DB_PASS}@tcp({DB_HOST})/{DB_NAME}?parseTime=True&charset={DB_CHARSET}&loc={DB_TZ}

	builder := strings.Builder{}

	builder.WriteString(self.User)
	builder.WriteString("@")

	if self.Pass != "" {
		builder.WriteString(self.Pass)
	}

	if self.Socket != "" {
		builder.WriteString(fmt.Sprintf("unix(%s)", self.Socket))
	} else if self.Host != "" {
		builder.WriteString(fmt.Sprintf("tcp(%s)", self.Host))
	}

	builder.WriteString("/%s?parseTime=True")

	if self.CharSet != "" {
		builder.WriteString("&charset=" + self.CharSet)
	}

	if self.TimeZone != "" {
		builder.WriteString("&loc=" + self.TimeZone)
	}

	s := builder.String()

	primaryDSN = fmt.Sprintf(s, self.Name)
	fallbackDSN = fmt.Sprintf(s, "")

	return
}
