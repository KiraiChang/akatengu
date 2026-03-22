package bootstrap

import (
	"os"
	"strconv"
	"time"
)

type Config struct {
	Env    string
	Server ServerConfig
	DB     DBConfig
	JWT    string
}

type ServerConfig struct {
	Addr         string // ":8080"
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
}

type DBConfig struct {
	Driver          string // sqlite / mysql / postgres
	DataSource      string
	Param           string
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
}

func LoadConfig() Config {
	return Config{
		Env: getEnv("APP_ENV", "local"),
		JWT: getEnv("JWT_SECRET", "this is a secret JWT"),

		Server: ServerConfig{
			Addr:         getEnv("SERVER_ADDR", ":8080"),
			ReadTimeout:  getDuration("SERVER_READ_TIMEOUT", 5*time.Second),
			WriteTimeout: getDuration("SERVER_WRITE_TIMEOUT", 10*time.Second),
		},

		DB: DBConfig{
			Driver:          getEnv("DB_DRIVER", "sqlite"),
			DataSource:      getEnv("DATA_SOURCE", "../output/akatengu/db.db"),
			Param:           getEnv("DB_PARAM", "?_foreign_keys=on&_busy_timeout=5000&_journal_mode=WAL&_synchronous=NORMAL"),
			MaxOpenConns:    getInt("DB_MAX_OPEN_CONNS", 1),
			MaxIdleConns:    getInt("DB_MAX_IDLE_CONNS", 1),
			ConnMaxLifetime: getDuration("DB_CONN_MAX_LIFETIME", time.Hour),
		},
	}
}

func getEnv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func getInt(key string, def int) int {
	if v := os.Getenv(key); v != "" {
		if i, err := strconv.Atoi(v); err == nil {
			return i
		}
	}
	return def
}

func getBool(key string, def bool) bool {
	if v := os.Getenv(key); v != "" {
		if b, err := strconv.ParseBool(v); err == nil {
			return b
		}
	}
	return def
}

func getDuration(key string, def time.Duration) time.Duration {
	if v := os.Getenv(key); v != "" {
		if d, err := time.ParseDuration(v); err == nil {
			return d
		}
	}
	return def
}
