package config

import "os"

type Config struct {
	Port      string
	DBName    string
	JWTSecret string
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

	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		jwtSecret = "какая то залупа 3000"
	}

	return &Config{Port: port, DBName: dbname, JWTSecret: jwtSecret}
}
