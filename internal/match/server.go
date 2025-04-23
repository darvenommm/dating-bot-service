package match

import (
	"encoding/json"
	"log"

	"github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/darvenommm/dating-bot-service/internal/action"
	"github.com/darvenommm/dating-bot-service/internal/broker"
	matchv1 "github.com/darvenommm/dating-bot-service/pkg/api/match/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type MatchServer struct {
	matchv1.UnimplementedMatchServiceServer
}

func NewServer() *MatchServer {
	return &MatchServer{}
}

func (s *MatchServer) ListenMatches(
	request *matchv1.ListenMatchesRequest,
	stream matchv1.MatchService_ListenMatchesServer,
) error {
	consumer, err := broker.CreateConsumer("match-group")
	if err != nil {
		return status.Errorf(codes.Internal, "kafka consumer: %v", err)
	}

	topic := "match-topic"
	if err := consumer.Subscribe(topic, nil); err != nil {
		return status.Errorf(codes.Internal, "subscribe: %v", err)
	}

	for {
		if stream.Context().Err() != nil {
			return nil
		}

		event := consumer.Poll(100)
		if event == nil {
			continue
		}

		switch msg := event.(type) {
		case *kafka.Message:
			var event action.MatchEvent
			if err := json.Unmarshal(msg.Value, &event); err != nil {
				log.Printf("bad message: %v", err)
				continue
			}

			match := matchv1.ListenMatchesResponse{
				FromUserId: event.FromUserId,
				ToUserId:   event.ToUserId,
			}

			if err := stream.Send(&match); err != nil {
				return err
			}

		case kafka.Error:
			log.Printf("kafka error: %v", msg)
		}
	}
}
