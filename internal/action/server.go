package action

import (
	"context"
	"encoding/json"

	"github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/darvenommm/dating-bot-service/internal/broker"
	"github.com/darvenommm/dating-bot-service/internal/orm"
	actionv1 "github.com/darvenommm/dating-bot-service/pkg/api/action/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type ActionServer struct {
	actionv1.UnimplementedActionServiceServer
	orm *orm.ORM
}

func NewServer(orm *orm.ORM) *ActionServer {
	return &ActionServer{orm: orm}
}

func (s *ActionServer) AddAction(
	_ context.Context,
	request *actionv1.AddActionRequest,
) (*actionv1.AddActionResponse, error) {
	topic := "action-topic"

	data, err := json.Marshal(request)
	if err != nil {
		return nil, err
	}

	producer, err := broker.CreateProducer()
	if err != nil {
		return nil, status.Errorf(codes.Internal, "create producer: %v", err)
	}

	err = producer.Produce(&kafka.Message{
		TopicPartition: kafka.TopicPartition{
			Topic:     &topic,
			Partition: kafka.PartitionAny,
		},
		Value: []byte(data),
	}, nil)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "producer: %v", err)
	}

	e := <-producer.Events()
	message := e.(*kafka.Message)

	if message.TopicPartition.Error != nil {
		return nil, status.Errorf(codes.Internal, "sending: %v", err)
	}

	return &actionv1.AddActionResponse{}, nil
}
