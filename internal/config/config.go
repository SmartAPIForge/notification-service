package config

import (
	"fmt"
	"github.com/joho/godotenv"
	"os"
	"strconv"
	"time"
)

type Config struct {
	Env               string // dev || prod
	RedisAddress      string
	RedisDb           int
	S3Path            string
	S3AccessKey       string
	S3SecretKey       string
	S3Bucket          string
	SchemaRegistryUrl string
	KafkaHost         string
}

type GRPCConfig struct {
	Port    int
	Timeout time.Duration
}

func MustLoad() *Config {
	loadEnvFile()

	env := getEnv("ENV", "dev")
	redisAddress := getEnv("REDIS_ADDRESS", "localhost:5252")
	redisDb := getEnvAsInt("REDIS_DB", 0)
	s3Path := mustGetEnv("S3_PATH")
	s3AccessKey := mustGetEnv("S3_ACCESS_KEY")
	s3SecretKey := mustGetEnv("S3_SECRET_KEY")
	s3Bucket := mustGetEnv("S3_BUCKET")
	schemaRegistryUrl := getEnv("SCHEMA_REGISTRY_URL", "http://localhost:6767")
	kafkaHost := getEnv("KAFKA_HOST", "http://localhost:9092")

	return &Config{
		Env:               env,
		RedisAddress:      redisAddress,
		RedisDb:           redisDb,
		S3Path:            s3Path,
		S3AccessKey:       s3AccessKey,
		S3SecretKey:       s3SecretKey,
		S3Bucket:          s3Bucket,
		SchemaRegistryUrl: schemaRegistryUrl,
		KafkaHost:         kafkaHost,
	}
}

func loadEnvFile() {
	if err := godotenv.Load(); err != nil {
		panic("Error loading .env file")
	}
}

func getEnv(key string, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}

func mustGetEnv(key string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	panic(fmt.Sprintf("Missed important variable in .env - %s", key))
}

func getEnvAsInt(key string, defaultValue int) int {
	valueStr := getEnv(key, "")
	if value, err := strconv.Atoi(valueStr); err == nil {
		return value
	}
	return defaultValue
}
