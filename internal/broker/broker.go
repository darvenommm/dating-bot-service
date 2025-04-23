package broker

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/confluentinc/confluent-kafka-go/kafka"
)

func getEnv(key string) (string, error) {
	if v, ok := os.LookupEnv(key); ok && v != "" {
		return v, nil
	}

	return "", fmt.Errorf("env variable %s is not set or empty", key)
}

func buildConfig() (*kafka.ConfigMap, error) {
	host, err := getEnv("KAFKA_ADVERTISED_HOST")
	if err != nil {
		return nil, err
	}

	port, err := getEnv("KAFKA_ADVERTISED_PORT")
	if err != nil {
		return nil, err
	}

	return &kafka.ConfigMap{
		"bootstrap.servers": fmt.Sprintf("%s:%s", host, port),
	}, nil
}

func CreateTopics(ctx context.Context) error {
	cfg, err := buildConfig()
	if err != nil {
		return err
	}

	admin, err := kafka.NewAdminClient(cfg)
	if err != nil {
		return err
	}
	defer admin.Close()

	specs := []kafka.TopicSpecification{
		{Topic: "match-topic", NumPartitions: 1, ReplicationFactor: 1},
		{Topic: "action-topic", NumPartitions: 1, ReplicationFactor: 1},
	}

	results, err := admin.CreateTopics(ctx, specs, kafka.SetAdminOperationTimeout(10*time.Second))
	if err != nil {
		return err
	}

	for _, r := range results {
		switch r.Error.Code() {
		case kafka.ErrNoError:
		case kafka.ErrTopicAlreadyExists:
			log.Printf("topic %q already exists, skipping", r.Topic)
		default:
			return fmt.Errorf("topic %s error: %v", r.Topic, r.Error)
		}
	}

	return nil
}

func CreateProducer() (*kafka.Producer, error) {
	cfg, err := buildConfig()
	if err != nil {
		return nil, err
	}

	return kafka.NewProducer(cfg)
}

func CreateConsumer(group string) (*kafka.Consumer, error) {
	cfg, err := buildConfig()
	if err != nil {
		return nil, err
	}

	if err := cfg.SetKey("group.id", group); err != nil {
		return nil, err
	}

	if err := cfg.SetKey("auto.offset.reset", "earliest"); err != nil {
		return nil, err
	}

	return kafka.NewConsumer(cfg)
}
