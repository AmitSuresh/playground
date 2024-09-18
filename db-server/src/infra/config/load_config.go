package config

import (
	"fmt"
	"os"

	"go.uber.org/zap"
)

func LoadConfig(l *zap.Logger) (*Config, error) {
	/* cwd, _ := os.Getwd()                  // Get current working directory
	envPath := filepath.Join(cwd, ".env") // Combine with .env filename
	err := godotenv.Load(envPath)
	if err != nil {
		l.Error("error loading .env file")
	} */
	config := &Config{
		DBipAddr:   os.Getenv("dbIP"),
		DBPort:     os.Getenv("dbPort"),
		DBUsername: os.Getenv("dbUser"),
		DBPassword: os.Getenv("dbPass"),
		DBName:     os.Getenv("dbName"),
		DBSSLMode:  os.Getenv("dbSSLMode"),
		VERSION:    os.Getenv("version"),
		ServerAddr: os.Getenv("serverAddr"),
		ServerPort: os.Getenv("serverPort"),
	}

	config.DBUrl = fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=%s",
		config.DBipAddr,
		config.DBUsername,
		config.DBPassword,
		config.DBName,
		config.DBPort,
		config.DBSSLMode)

	return config, nil
}
