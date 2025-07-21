package config

import (
	"bufio"
	"embed"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"
	"time"
)

//go:embed .env.*
var envFiles embed.FS

type HttpServerConfig struct {
	Protocol     string
	Host         string
	Port         string
	IdleTimeout  time.Duration
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
}

func (h HttpServerConfig) URL() string {
	return fmt.Sprintf("%s://%s:%s", h.Protocol, h.Host, h.Port)
}

type DBConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	Name     string
}

type KafkaConfig struct {
	Addr          string
	ProducerTopic string
	ConsumerTopic string
	GroupID       string
}

type Config struct {
	Env             string
	GracefulTimeout time.Duration

	Inventory struct {
		HTTP  HttpServerConfig
		DB    DBConfig
		Kafka KafkaConfig
	}

	Order struct {
		HTTP  HttpServerConfig
		DB    DBConfig
		Kafka KafkaConfig
	}

	Storefront struct {
		HTTP HttpServerConfig
	}
}

func Load() (*Config, error) {
	if err := loadEnvFileFromEmbed(".env.common"); err != nil {
		return nil, err
	}

	envName := os.Getenv("GO_ENV")
	if envName == "" {
		envName = "dev"
	}
	if err := loadEnvFileFromEmbed(".env." + envName); err != nil {
		return nil, err
	}

	cfg := &Config{
		Env:             getEnv("GO_ENV", "dev"),
		GracefulTimeout: getEnvDuration("GRACEFUL_TIMEOUT", 2*time.Second),
		Inventory: struct {
			HTTP  HttpServerConfig
			DB    DBConfig
			Kafka KafkaConfig
		}{
			HTTP: HttpServerConfig{
				Host:         getEnv("INVENTORY_HOST", "localhost"),
				Port:         getEnv("INVENTORY_PORT", "8081"),
				Protocol:     getEnv("INVENTORY_PROTOCOL", "http"),
				ReadTimeout:  getEnvDuration("INVENTORY_READ_TIMEOUT", 10*time.Second),
				WriteTimeout: getEnvDuration("INVENTORY_WRITE_TIMEOUT", 10*time.Second),
				IdleTimeout:  getEnvDuration("INVENTORY_IDLE_TIMEOUT", 10*time.Second),
			},
			DB: DBConfig{
				Host:     getEnv("INVENTORY_DB_HOST", "localhost"),
				Port:     getEnv("INVENTORY_DB_PORT", "5432"),
				User:     getEnv("INVENTORY_DB_USER", "postgres"),
				Password: getEnv("INVENTORY_DB_PASSWORD", "postgres"),
				Name:     getEnv("INVENTORY_DB_NAME", "postgres"),
			},
			Kafka: KafkaConfig{
				Addr:          getEnv("INVENTORY_KAFKA_ADDR", "localhost:9092"),
				ProducerTopic: getEnv("INVENTORY_KAFKA_PRODUCER_TOPIC", "inventory.reserved"),
				ConsumerTopic: getEnv("INVENTORY_KAFKA_CONSUMER_TOPIC", "orders.created,payments.failed"),
				GroupID:       getEnv("INVENTORY_KAFKA_GROUP_ID", "inventory-service"),
			},
		},
		Order: struct {
			HTTP  HttpServerConfig
			DB    DBConfig
			Kafka KafkaConfig
		}{
			HTTP: HttpServerConfig{
				Host:         getEnv("ORDER_HOST", "localhost"),
				Port:         getEnv("ORDER_PORT", "8082"),
				Protocol:     getEnv("ORDER_PROTOCOL", "http"),
				ReadTimeout:  getEnvDuration("ORDER_READ_TIMEOUT", 10*time.Second),
				WriteTimeout: getEnvDuration("ORDER_WRITE_TIMEOUT", 10*time.Second),
				IdleTimeout:  getEnvDuration("ORDER_IDLE_TIMEOUT", 10*time.Second),
			},
			DB: DBConfig{
				Host:     getEnv("ORDER_DB_HOST", "localhost"),
				Port:     getEnv("ORDER_DB_PORT", "5432"),
				User:     getEnv("ORDER_DB_USER", "postgres"),
				Password: getEnv("ORDER_DB_PASSWORD", "postgres"),
				Name:     getEnv("ORDER_DB_NAME", "postgres"),
			},
			Kafka: KafkaConfig{
				Addr:          getEnv("ORDER_KAFKA_BROKER", "localhost:9092"),
				ProducerTopic: getEnv("ORDER_KAFKA_WRITER_TOPIC", "order-events"),
				ConsumerTopic: getEnv("ORDER_KAFKA_READER_TOPIC", "inventory-events,payment-events"),
				GroupID:       getEnv("ORDER_KAFKA_GROUP_ID", "order-service-group"),
			},
		},
		Storefront: struct {
			HTTP HttpServerConfig
		}{
			HTTP: HttpServerConfig{
				Host:         getEnv("STOREFRONT_HOST", "localhost"),
				Port:         getEnv("STOREFRONT_PORT", "8080"),
				Protocol:     getEnv("STOREFRONT_PROTOCOL", "http"),
				ReadTimeout:  getEnvDuration("STOREFRONT_READ_TIMEOUT", 10*time.Second),
				WriteTimeout: getEnvDuration("STOREFRONT_WRITE_TIMEOUT", 10*time.Second),
				IdleTimeout:  getEnvDuration("STOREFRONT_IDLE_TIMEOUT", 10*time.Second),
			},
		},
	}

	log.Printf("Config loaded: %v", prettyPrint(cfg))
	return cfg, nil
}

// loadEnvFileFromEmbed loads env file from the embedded FS
func loadEnvFileFromEmbed(filename string) error {
	f, err := envFiles.Open(filename)
	if err != nil {
		return err
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}
		key := strings.TrimSpace(parts[0])
		value := strings.Trim(strings.TrimSpace(parts[1]), `"`)
		os.Setenv(key, value)
	}
	return nil
}

func getEnv(key, defaultVal string) string {
	if val, exists := os.LookupEnv(key); exists {
		return val
	}
	return defaultVal
}

func getEnvDuration(key string, defaultVal time.Duration) time.Duration {
	if valStr, ok := os.LookupEnv(key); ok {
		if d, err := time.ParseDuration(valStr); err == nil {
			return d
		}
	}
	return defaultVal
}

func prettyPrint(v interface{}) string {
	b, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return fmt.Sprintln("error:", err)
	} else {
		return fmt.Sprintln(string(b))
	}
}
