package config_test

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"interview/pkg/config"
	"os"
	"testing"
)

func TestLoadConfig(t *testing.T) {

	t.Run("load config", func(t *testing.T) {
		f, err := os.CreateTemp(".", "*.yml")
		require.NoError(t, err)
		defer os.Remove(f.Name())
		f.WriteString(`
server:
  host:
  port: 8080

database:
  name: ice_db_test
  type: mysql
  username: test
  password: test
  host: localhost
  port: 4001
`)
		envFile, err := os.CreateTemp(".", "*.env")
		require.NoError(t, err)
		defer os.Remove(envFile.Name())
		envFile.WriteString("DB_PASSWORD= mysecretpassword")

		cfg, err := config.LoadConfig(f.Name(), envFile.Name())
		assert.NoError(t, err)
		assert.NotEmpty(t, cfg)
		assert.Equal(t, "8080", cfg.Server.Port)
		assert.Equal(t, "ice_db_test", cfg.Database.Name)
		assert.Equal(t, "mysecretpassword", cfg.Database.Password)
	})

	t.Run("invalid config path", func(t *testing.T) {
		_, err := config.LoadConfig("some-path")
		assert.Error(t, err)
	})

	t.Run("invalid env file path", func(t *testing.T) {
		f, err := os.CreateTemp(".", "*.yml")
		defer os.Remove(f.Name())

		_, err = config.LoadConfig(f.Name(), "path/to/.env")
		assert.Error(t, err)
	})
}
