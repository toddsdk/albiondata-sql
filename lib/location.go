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

	SwampCrossMarket    Location = 4
	ForestCrossMarket   Location = 1006
	SteppeCrossMarket   Location = 2002
	HighlandCrossMarket Location = 3002
	MountainCrossMarket Location = 4006

	// SwampOutpostMarket Location = 0004#1
	// ForestOutpostMarket Location = 1006#1
	// SteppeOutpostMarket Location = 2002#1
	// HighlandOutpostMarket Location = 3002#1
	// MountainOutpostMarket Location = 4006#1
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
