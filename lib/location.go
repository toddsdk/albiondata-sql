package lib

import (
	"fmt"
)

type Location int

const (
	Unknown            Location = 0
	ThetfordMarket     Location = 7
	LymhurstMarket     Location = 1002
	BridgewatchMarket  Location = 2004
	CaerleonMarket     Location = 3005
	MartlockMarket     Location = 3010
	FortSterlingMarket Location = 4002
)

func Locations() []Location {
	return []Location{
		ThetfordMarket,
		LymhurstMarket,
		BridgewatchMarket,
		CaerleonMarket,
		MartlockMarket,
		FortSterlingMarket,
	}
}

func (l Location) TableName() string {
	switch int(l) {
	case int(ThetfordMarket):
		return "thetford_market_orders"
	case int(LymhurstMarket):
		return "lymhurst_market_orders"
	case int(BridgewatchMarket):
		return "bridgewatch_market_orders"
	case int(CaerleonMarket):
		return "caerleon_market_orders"
	case int(MartlockMarket):
		return "martlock_market_orders"
	case int(FortSterlingMarket):
		return "fort_sterling_market_orders"
	default:
		// Will never happen
		return ""
	}
}

func (l Location) String() string {
	switch int(l) {
	case int(ThetfordMarket):
		return "Thetford Market"
	case int(LymhurstMarket):
		return "Lymhurst Market"
	case int(BridgewatchMarket):
		return "Bridgewatch Market"
	case int(CaerleonMarket):
		return "Caerleon Market"
	case int(MartlockMarket):
		return "Martlock Market"
	case int(FortSterlingMarket):
		return "Fort Sterling Market"
	default:
		// Will never happen
		return ""
	}
}

func (l Location) Model() ModelMarketOrder {
	return NewModelMarketOrder(l)
}

func NewLocationFromId(locationID int) (Location, error) {
	switch locationID {
	case int(ThetfordMarket):
		return ThetfordMarket, nil
	case int(LymhurstMarket):
		return LymhurstMarket, nil
	case int(BridgewatchMarket):
		return BridgewatchMarket, nil
	case int(CaerleonMarket):
		return CaerleonMarket, nil
	case int(MartlockMarket):
		return MartlockMarket, nil
	case int(FortSterlingMarket):
		return FortSterlingMarket, nil
	default:
		return Unknown, fmt.Errorf("Unknown location: %d", locationID)
	}
}
