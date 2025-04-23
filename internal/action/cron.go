package action

import (
	"encoding/json"
	"log"
	"time"

	"github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/darvenommm/dating-bot-service/internal/broker"
	"github.com/darvenommm/dating-bot-service/internal/orm"
)

type MatchEvent struct {
	FromUserId int64 `json:"from_user_id"`
	ToUserId   int64 `json:"to_user_id"`
}

func CheckMatchesCron(orm *orm.ORM) {
	go func() {
		ticker := time.NewTicker(10 * time.Second)
		defer ticker.Stop()

		log.Println("match cron started")

		for {
			<-ticker.C
			log.Println("checking for matches (self join)...")

			type MatchPair struct {
				ID1         uint
				ID2         uint
				FromUserID1 int
				ToUserID1   int
			}

			var pairs []MatchPair
			err := orm.DB().
				Raw(`
				SELECT
					a.id AS id1, b.id AS id2,
					a.from_user_id AS from_user_id1,
					a.to_user_id AS to_user_id1
				FROM user_actions a
				JOIN user_actions b
					ON a.from_user_id = b.to_user_id
					AND a.to_user_id = b.from_user_id
					AND a.action = ?
					AND b.action = ?
				WHERE a.was_matched = false
					AND b.was_matched = false
					AND a.from_user_id < b.from_user_id
			`, Like, Like).
				Scan(&pairs).Error

			if err != nil {
				log.Printf("query error: %v", err)
				continue
			}

			if len(pairs) == 0 {
				log.Println("no new matches found")
				continue
			}

			producer, err := broker.CreateProducer()
			if err != nil {
				log.Printf("kafka producer error: %v", err)
				continue
			}

			for _, pair := range pairs {
				err := orm.DB().Model(&UserAction{}).
					Where("id IN ?", []uint{pair.ID1, pair.ID2}).
					Update("was_matched", true).Error

				if err != nil {
					log.Printf("failed to update match [%d, %d]: %v", pair.ID1, pair.ID2, err)
					continue
				}

				event := MatchEvent{
					FromUserId: int64(pair.FromUserID1),
					ToUserId:   int64(pair.ToUserID1),
				}
				data, err := json.Marshal(event)
				if err != nil {
					log.Printf("failed to marshal match event: %v", err)
					continue
				}

				topic := "match-topic"
				err = producer.Produce(&kafka.Message{
					TopicPartition: kafka.TopicPartition{
						Topic:     &topic,
						Partition: kafka.PartitionAny,
					},
					Value: data,
				}, nil)

				if err != nil {
					log.Printf("failed to send Kafka message: %v", err)
				} else {
					log.Printf("sent match to Kafka: %+v", event)
				}
			}

			producer.Close()
		}
	}()
}
