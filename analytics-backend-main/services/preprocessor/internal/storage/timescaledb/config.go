// preprocessor/internal/storage/timescaledb/config.go

package timescaledb

import "fmt"

// Config описывает настройки подключения и миграций к TimescaleDB.
type Config struct {
	DSN           string `mapstructure:"dsn"`
	MigrationsDir string `mapstructure:"migrations_dir"`
}

func (c *Config) ApplyDefaults() {
	if c.MigrationsDir == "" {
		c.MigrationsDir = "/app/migrations/timescaledb"
	}
}

func (c *Config) Validate() error {
	if c.DSN == "" {
		return fmt.Errorf("timescaledb: dsn must be provided")
	}
	if c.MigrationsDir == "" {
		return fmt.Errorf("timescaledb: migrations_dir must be provided")
	}
	return nil
}
