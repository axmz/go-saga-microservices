package config

import (
	"embed"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

//go:embed config.yaml
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
	// Open the embedded YAML config file
	f, err := envFiles.Open("config.yaml")
	if err != nil {
		return nil, err
	}
	defer f.Close()

	// Load the whole YAML into a map for merging
	var layered map[string]map[string]interface{}
	if err := yaml.NewDecoder(f).Decode(&layered); err != nil {
		return nil, err
	}

	envName := os.Getenv("GO_ENV")
	if envName == "" {
		envName = "dev"
	}

	// Merge common and env-specific configs
	merged := deepCopyMap(layered["common"])
	if envSection, ok := layered[envName]; ok {
		mergeMaps(merged, envSection)
	}

	// Marshal merged map back to YAML, then unmarshal into Config struct
	mergedYAML, err := yaml.Marshal(merged)
	if err != nil {
		return nil, err
	}

	var cfg Config
	if err := yaml.Unmarshal(mergedYAML, &cfg); err != nil {
		return nil, err
	}

	log.Printf("Config loaded: %v", prettyPrint(cfg))
	return &cfg, nil
}

func prettyPrint(v interface{}) string {
	b, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return fmt.Sprintln("error:", err)
	} else {
		return fmt.Sprintln(string(b))
	}
}

// deepCopyMap creates a deep copy of a map[string]interface{}
func deepCopyMap(src map[string]interface{}) map[string]interface{} {
	copy := make(map[string]interface{}, len(src))
	for k, v := range src {
		if m, ok := v.(map[string]interface{}); ok {
			copy[k] = deepCopyMap(m)
		} else {
			copy[k] = v
		}
	}
	return copy
}

// mergeMaps overlays src into dst recursively
func mergeMaps(dst, src map[string]interface{}) {
	for k, v := range src {
		if vMap, ok := v.(map[string]interface{}); ok {
			if dstMap, ok := dst[k].(map[string]interface{}); ok {
				mergeMaps(dstMap, vMap)
			} else {
				dst[k] = deepCopyMap(vMap)
			}
		} else {
			dst[k] = v
		}
	}
}
