package models

import (
	"github.com/lib/pq"
	"github.com/pgvector/pgvector-go"
	"gorm.io/gorm"
)

type Post struct {
	gorm.Model

	KonachanID int    `gorm:"uniqueIndex;not null"`
	ImageURL   string `gorm:"not null"`
	PreviewURL string `gorm:"not null"`

	Tags     pq.StringArray `gorm:"type:text[];index:,type:gin"`
	Rating   string         `gorm:"type:varchar(1);index"`
	Width    int
	Height   int
	Score    int
	FileSize int

	Temperature float64     `gorm:"type:float"`
	Palette     []PostColor `gorm:"foreignKey:PostID"`
}

type PostColor struct {
	ID     uint
	PostID uint
	Color  pgvector.Vector `gorm:"type:vector(3)"`
	Weight float64         `gorm:"type:float"`
}
