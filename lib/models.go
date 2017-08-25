package lib

import (
	"time"

	"github.com/jinzhu/gorm"
)

type ModelMarketOrder struct {
	AlbionID         uint   `gorm:"not null;unique_index:idx_id_location"`
	ItemID           string `gorm:"index"`
	QualityLevel     int8   `gorm:"size:1"`
	EnchantmentLevel int8   `gorm:"size:1"`
	Price            int    `gorm:"index"`
	InitialAmount    int
	Amount           int
	AuctionType      string `gorm:"index"`
	Expires          time.Time
	Location         Location `gorm:"not null;unique_index:idx_id_location;index"`
	gorm.Model
}

func (m ModelMarketOrder) TableName() string {
	return "market_orders"
}

func NewModelMarketOrder() ModelMarketOrder {
	return ModelMarketOrder{}
}
