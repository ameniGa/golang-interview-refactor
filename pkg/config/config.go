package config

import (
	"github.com/joho/godotenv"
	"github.com/spf13/viper"
	"path/filepath"
	"runtime"
)

type Server struct {
	Host string `mapstructure:"host"`
	Port string `mapstructure:"port"`
}

// Database database configuration.
type Database struct {
	Name     string `mapstructure:"name"`
	Type     string `mapstructure:"type"`
	Username string `mapstructure:"username"`
	Password string `mapstructure:"password"`
	Host     string `mapstructure:"host"`
	Port     string `mapstructure:"port"`
}

type Config struct {
	Server   Server   `yaml:"server"`
	Database Database `yaml:"database"`
}

// LoadConfig loads the configuration
func LoadConfig(path string, envFile ...string) (config Config, err error) {
	viper.Reset()
	dir, file := filepath.Split(path)
	_, filename, _, _ := runtime.Caller(0)

	viper.AddConfigPath(filepath.Join(filepath.Dir(filename), "../config"))
	viper.AddConfigPath(dir)
	viper.SetConfigName(file)
	viper.SetConfigType("yaml")

	if envFile != nil {
		err = godotenv.Load(envFile...)
		if err != nil {
			return
		}
	}

	// allow defaulting the keys to environment variable values
	viper.AutomaticEnv()
	// read the config file(s)
	err = viper.ReadInConfig()
	if err != nil {
		return
	}

	_ = viper.BindEnv("database.password", "DB_PASSWORD")
	_ = viper.BindEnv("database.username", "DB_USERNAME")

	// parse the configuration
	err = viper.Unmarshal(&config)
	return
}
