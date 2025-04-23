package profile

import "gorm.io/gorm"

type Gender uint

const (
	Male Gender = iota
	Female
)

type Profile struct {
	gorm.Model
	UserID           int    `gorm:"not null;index"`
	FullName         string `gorm:"size:255;not null"`
	Gender           Gender `gorm:"not null"`
	Age              uint   `gorm:"not null;check:age>=0"`
	Description      string `gorm:"type:text"`
	Photo            []byte `gorm:"type:bytea"`
	PrimaryRating    int    `gorm:"default:0;check:primary_rating>=30"`
	BehavioralRating int    `gorm:"default:0;check:behavioral_rating>=0"`
	ResultRating     int    `gorm:"->;type:integer GENERATED ALWAYS AS ((primary_rating + behavioral_rating) / 2) STORED"`
}

func (p *Profile) BeforeSave(tx *gorm.DB) error {
	hasDescription := p.Description != ""
	hasPhoto := len(p.Photo) > 0

	switch {
	case hasDescription && hasPhoto:
		p.PrimaryRating = 100
	case hasDescription || hasPhoto:
		p.PrimaryRating = 60
	default:
		p.PrimaryRating = 30
	}

	return nil
}

func Migrate(db *gorm.DB) error {
	return db.AutoMigrate(&Profile{})
}
