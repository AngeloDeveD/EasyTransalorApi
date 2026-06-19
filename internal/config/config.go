package config

import "os"

type Config struct {
	Port   string
	DBName string
}

func Load() *Config {
	port := os.Getenv("APP_PORT")
	if port == "" {
		port = "8080"
	}

	dbname := os.Getenv("DB_NAME")
	if dbname == "" {
		dbname = "app.db"
	}

	return &Config{Port: port, DBName: dbname}
}
