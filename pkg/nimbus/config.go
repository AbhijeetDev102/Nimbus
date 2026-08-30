package nimbus

import (
	"fmt"
	"os"

	"github.com/google/uuid"
)

// Config holds all infrastructure connection settings for a Nimbus Worker
type Config struct {
	WorkerID     uuid.UUID
	PostgresDSN  string
	KafkaBrokers []string
	KafkaGroup   string
	KafkaTopic   string
	RedisAddr    string
}

func getEnv(key, defaultVal string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultVal
}

// ConfigFromEnv reads standard environment variables into Config
func ConfigFromEnv() Config {
	workerID := uuid.New()
	host := getEnv("POSTGRES_HOSTNAME", "localhost")
	user := getEnv("POSTGRES_USER", "postgres")
	password := getEnv("POSTGRES_PASSWORD", "postgrespassword")
	dbName := getEnv("POSTGRES_DB_NAME", "nimbus")
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=5432 sslmode=disable", host, user, password, dbName)
	return Config{
		WorkerID:     workerID,
		PostgresDSN:  dsn,
		KafkaBrokers: []string{getEnv("KAFKA_BROKERS", "localhost:9092")},
		KafkaGroup:   getEnv("KAFKA_CONSUMER_GROUP", "nimbus-worker-group"),
		KafkaTopic:   getEnv("KAFKA_JOB_TOPIC", "job.events"),
		RedisAddr:    getEnv("REDIS_ADDR", "localhost:6379"),
	}
}
