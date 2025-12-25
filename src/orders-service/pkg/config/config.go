package config

import "os"

type Config struct {
	Database DatabaseConfig
	RabbitMQ RabbitMQConfig
	Server   ServerConfig
}

type DatabaseConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	Name     string
}

type RabbitMQConfig struct {
	Host     string
	User     string
	Password string
}

type ServerConfig struct {
	Port string
}

func Load() *Config {
	return &Config{
		Database: DatabaseConfig{
			Host:     getEnv("DB_HOST", "localhost"),
			Port:     getEnv("DB_PORT", "5432"),
			User:     getEnv("DB_USER", "orders_user"),
			Password: getEnv("DB_PASSWORD", "orders_pass"),
			Name:     getEnv("DB_NAME", "orders_db"),
		},
		RabbitMQ: RabbitMQConfig{
			Host:     getEnv("RABBITMQ_HOST", "localhost"),
			User:     getEnv("RABBITMQ_USER", "guest"),
			Password: getEnv("RABBITMQ_PASS", "guest"),
		},
		Server: ServerConfig{
			Port: getEnv("SERVER_PORT", "8080"),
		},
	}
}

func (c *Config) GetDSN() string {
	return "host=" + c.Database.Host +
		" user=" + c.Database.User +
		" password=" + c.Database.Password +
		" dbname=" + c.Database.Name +
		" port=" + c.Database.Port +
		" sslmode=disable"
}

func (c *Config) GetRabbitMQURL() string {
	return "amqp://" + c.RabbitMQ.User + ":" + c.RabbitMQ.Password + "@" + c.RabbitMQ.Host + ":5672/"
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
