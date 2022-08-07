package crawldb

import (
	"time"

	"github.com/JOwlHouse/onionscan/pkg/model"
	"gorm.io/gorm"
)

// CrawlRecord defines a spider entry in the database
type CrawlRecord struct {
	gorm.Model

	URL       string `gorm:"primarykey"`
	Timestamp time.Time
	Page      model.Page
	RShips    []Relationship
}

// Relationship defines a correltion record in the Database.
type Relationship struct {
	gorm.Model

	ID         int `gorm:"primarykey"`
	Onion      string
	From       string
	Type       string
	Identifier string
	FirstSeen  time.Time
	LastSeen   time.Time
}
