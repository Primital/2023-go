package main

// MapData types
type (
	MapData struct {
		Name      string              `json:"mapName"`
		Border    Border              `json:"border"`
		Locations map[string]Location `json:"locations"`
		HotSpots  []HotSpot           `json:"hotSpots"`
		TypeCount LocationTypeCount   `json:"locationTypeCount"`
	}

	Border struct {
		LatitudeMax  float64 `json:"latitudeMax"`
		LatitudeMin  float64 `json:"latitudeMin"`
		LongitudeMax float64 `json:"longitudeMax"`
		LongitudeMin float64 `json:"longitudeMin"`
	}

	Location struct {
		Name              string  `json:"locationName"`
		Type              string  `json:"locationType"`
		Latitude          float64 `json:"latitude"`
		Longitude         float64 `json:"longitude"`
		Footfall          float64 `json:"footfall"`
		FootfallScale     float64 `json:"footfallScale"`
		SalesVolume       float64 `json:"salesVolume"`
		neighborDistances map[string]float64
	}

	HotSpot struct {
		Spread    float64 `json:"spread"`
		Name      string  `json:"hotSpotName"`
		Latitude  float64 `json:"latitude"`
		Longitude float64 `json:"longitude"`
		Footfall  float64 `json:"footfall"`
	}

	LocationTypeCount struct {
		Total          int `json:"total"`
		GasStation     int `json:"gasStation"`
		GroceryStore   int `json:"groceryStore"`
		GroceryStoreLg int `json:"groceryStoreLg"`
		Kiosk          int `json:"kiosk"`
		Convenience    int `json:"convenience"`
	}
)

// GeneralGameData types
type (
	GeneralGameData struct {
		ClassicUnitData                 UnitData       `json:"classicUnitData"`
		RefillUnitData                  UnitData       `json:"refillUnitData"`
		Freestyle9100Data               RefillUnitData `json:"freestyle9100Data"`
		Freestyle3100Data               RefillUnitData `json:"freestyle3100Data"`
		LocationTypes                   LocationTypes  `json:"locationTypes"`
		CompetitionMapNames             []string       `json:"competitionMapNames"`
		TrainingMapNames                []string       `json:"trainingMapNames"`
		Co2PricePerKiloInSek            float64        `json:"co2PricePerKiloInSek"`
		WillingnessToTravelInMeters     float64        `json:"willingnessToTravelInMeters"`
		ConstantExpDistributionFunction float64        `json:"constantExpDistributionFunction"`
		RefillSalesFactor               float64        `json:"refillSalesFactor"`
		RefillDistributionRate          float64        `json:"refillDistributionRate"`
	}

	UnitData struct {
		Type              string  `json:"type"`
		Co2PerUnitInGrams float64 `json:"co2PerUnitInGrams"`
		ProfitPerUnit     float64 `json:"profitPerUnit"`
	}

	RefillUnitData struct {
		Type                  string  `json:"type"`
		LeasingCostPerWeek    float64 `json:"leasingCostPerWeek"`
		RefillCapacityPerWeek float64 `json:"refillCapacityPerWeek"`
		StaticCo2             float64 `json:"staticCo2"`
	}

	LocationTypes struct {
		GroceryStoreLarge LocationType `json:"groceryStoreLarge"`
		GroceryStore      LocationType `json:"groceryStore"`
		Convenience       LocationType `json:"convenience"`
		GasStation        LocationType `json:"gasStation"`
		Kiosk             LocationType `json:"kiosk"`
	}

	LocationType struct {
		Type        string  `json:"type"`
		SalesVolume float64 `json:"salesVolume"`
	}
)

type GameData struct {
	ID    string    `json:"id"`
	Score GameScore `json:"gameScore"`
}

type GameScore struct {
	Co2Savings    float64 `json:"kgCo2Savings"`
	Earnings      float64 `json:"earnings"`
	TotalFootfall float64 `json:"totalFootfall"`
	Total         float64 `json:"total"`
}

func (l Location) GetLocationsWithinWalkingDistance(locations map[string]LocationSolution, data GeneralGameData) map[string]float64 {
	lat, long := l.Latitude, l.Longitude
	locs := make(map[string]float64)
	for _, loc := range locations {
		if loc.Location.Name == l.Name {
			continue
		}
		if hav := haversine(lat, long, loc.Location.Latitude, loc.Location.Longitude); hav <= data.WillingnessToTravelInMeters {
			locs[loc.Location.Name] = hav
		}
	}
	return locs
}

func PrecalculateNeighborDistances(locations []*Location) {
	for _, loc := range locations {
		loc.neighborDistances = make(map[string]float64)
		for _, other := range locations {
			if loc.Name == other.Name {
				continue
			}
			dist := haversine(loc.Latitude, loc.Longitude, other.Latitude, other.Longitude)
			if dist <= 150.0 {
				loc.neighborDistances[other.Name] = dist
			}
		}
	}
}
