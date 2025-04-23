package filter

import "gorm.io/gorm"

type Gender uint

const (
	Male Gender = iota
	Female
)

type Filter struct {
	gorm.Model
	UserID int    `gorm:"not null"`
	Gender Gender `gorm:"not null"`
	MinAge uint   `gorm:"not null"`
	MaxAge uint   `gorm:"not null;check:max_age>=min_age"`
}

func Migrate(db *gorm.DB) error {
	return db.AutoMigrate(&Filter{})
}
