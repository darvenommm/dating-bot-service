package action

import "gorm.io/gorm"

type Action uint

const (
	Like Action = iota + 1
	Dislike
)

type UserAction struct {
	gorm.Model

	FromUserID int    `gorm:"not null;index"`
	ToUserID   int    `gorm:"not null;index"`
	Action     Action `gorm:"not null"`
	WasMatched bool   `gorm:"default:false"`
}

func Migrate(db *gorm.DB) error {
	return db.AutoMigrate(&UserAction{})
}
