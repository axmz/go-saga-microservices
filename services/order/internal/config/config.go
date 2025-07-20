package config

import (
	"log"
	"os"
	"strconv"
	"time"

	"github.com/axmz/go-saga-microservices/lib/adapter/db"
	"github.com/axmz/go-saga-microservices/lib/adapter/http"
	"github.com/axmz/go-saga-microservices/lib/adapter/kafka"
)

type Config struct {
	Env             string
	GracefulTimeout time.Duration
	HTTP            http.Config
	DB              db.Config
	Kafka           kafka.Config
}

type DB struct {
	Host     string
	Port     string
	User     string
	Password string
	Name     string
}

type HTTPServer struct {
	Protocol     string
	Host         string
	Port         string
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
	IdleTimeout  time.Duration
}

func MustLoad() (*Config, error) {
	cfg := &Config{
		Env:             getEnv("APP_ENV", "local"),
		GracefulTimeout: getEnvAsDuration("GRACEFUL_TIMEOUT", 2*time.Second),
		HTTP: http.Config{
			Protocol:     getEnv("PROTOCOL", "http"),
			Host:         getEnv("HOST", "0.0.0.0"),
			Port:         getEnv("PORT", "8080"),
			ReadTimeout:  getEnvAsDuration("READ_TIMEOUT", 5*time.Second),
			WriteTimeout: getEnvAsDuration("WRITE_TIMEOUT", 10*time.Second),
			IdleTimeout:  getEnvAsDuration("IDLE_TIMEOUT", 120*time.Second),
		},
		DB: db.Config{
			Host:     getEnv("DB_HOST", "localhost"),
			Port:     getEnv("DB_PORT", "5432"),
			User:     getEnv("DB_USER", "postgres"),
			Password: getEnv("DB_PASSWORD", "postgres"),
			Name:     getEnv("DB_NAME", "postgres"),
		},
		Kafka: kafka.Config{
			Addr:          getEnv("KAFKA_BROKER", "localhost:9092"),
			ProducerTopic: getEnv("KAFKA_WRITER_TOPIC", "order-events"),
			ConsumerTopic: getEnv("KAFKA_READER_TOPIC", "inventory-events,payment-events"),
			GroupID:       getEnv("KAFKA_GROUP_ID", "order-service-group"),
		},
	}

	log.SetOutput(os.Stdout)
	log.Printf("Loaded configuration: Env=%s, HTTP=%+v, DB=%+v, Kafka=%+v, GracefulTimeout=%v",
		cfg.Env, cfg.HTTP, cfg.DB, cfg.Kafka, cfg.GracefulTimeout)
	return cfg, nil
}

func getEnv(key, defaultVal string) string {
	if val, exists := os.LookupEnv(key); exists {
		return val
	}
	return defaultVal
}

func getEnvAsDuration(key string, defaultVal time.Duration) time.Duration {
	if valStr, ok := os.LookupEnv(key); ok {
		if valInt, err := strconv.Atoi(valStr); err == nil {
			return time.Duration(valInt) * time.Second
		} else {
			log.Fatalf("Invalid duration for %s: %v. Using default: %v", key, err, defaultVal)
		}
	}
	return defaultVal
}
