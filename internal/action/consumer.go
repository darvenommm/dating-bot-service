package action

import (
	"encoding/json"
	"log"

	"github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/darvenommm/dating-bot-service/internal/broker"
	"github.com/darvenommm/dating-bot-service/internal/orm"
	"github.com/darvenommm/dating-bot-service/internal/profile"
	actionv1 "github.com/darvenommm/dating-bot-service/pkg/api/action/v1"
	"gorm.io/gorm"
)

func StartListeningAction(db *orm.ORM) error {
	consumer, err := broker.CreateConsumer("action-group")
	if err != nil {
		return err
	}

	err = consumer.Subscribe("action-topic", nil)
	if err != nil {
		return err
	}

	go func() {
		for {
			ev := consumer.Poll(100)
			if ev == nil {
				continue
			}

			switch message := ev.(type) {
			case *kafka.Message:
				var request actionv1.AddActionRequest
				if err := json.Unmarshal(message.Value, &request); err != nil {
					continue
				}

				err := db.DB().Transaction(func(tx *gorm.DB) error {
					log.Println(request.GetFromUserId(), request.GetToUserId())
					action := UserAction{
						FromUserID: int(request.GetFromUserId()),
						ToUserID:   int(request.GetToUserId()),
						Action:     Action(request.GetAction()),
					}

					if err := tx.Create(&action).Error; err != nil {
						return err
					}

					var prof profile.Profile
					if err := tx.First(&prof, "user_id = ?", request.GetToUserId()).Error; err != nil {
						return err
					}

					switch request.GetAction() {
					case actionv1.Action_ACTION_LIKE:
						prof.BehavioralRating = min(prof.BehavioralRating+1, 100)
					case actionv1.Action_ACTION_DISLIKE:
						prof.BehavioralRating = max(prof.BehavioralRating-1, 0)
					}

					if err := tx.Save(&prof).Error; err != nil {
						return err
					}

					return nil
				})

				if err != nil {
					log.Printf("failed to process action: %v", err)
				}

			case kafka.Error:
				log.Printf("⚠ Kafka error: %v", message)
			default:
			}
		}
	}()

	return nil
}
