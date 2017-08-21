package lib

import (
	"time"
)

type ModelMarketOrder struct {
	ID               uint   `gorm:"primary_key"`
	ItemID           string `gorm:"index"`
	QualityLevel     int8   `gorm:"size:1"`
	EnchantmentLevel int8   `gorm:"size:1"`
	Price            int    `gorm:"index"`
	Amount           int
	AuctionType      string `gorm:"index"`
	Expires          time.Time
	CreatedAt        time.Time
	UpdatedAt        time.Time
	DeletedAt        *time.Time

	Location Location `gorm:"-"`
}

func (m *ModelMarketOrder) TableName() string {
	return m.Location.TableName()
}

func NewModelMarketOrder(location Location) *ModelMarketOrder {
	return &ModelMarketOrder{
		Location: location,
	}
}
