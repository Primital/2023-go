package scoring

import (
	"fmt"
	"math"

	"2023-go/internal"
	"2023-go/types"
)

type ScoredSolution struct {
	GameId           string                            `json:"gameId"`
	MapName          string                            `json:"mapName"`
	Locations        map[string]types.LocationSolution `json:"locations"`
	GameScore        map[string]float64                `json:"gameScore"`
	TotalRevenue     float64                           `json:"totalRevenue"`
	TotalLeasingCost float64                           `json:"totalLeasingCost"`
	TotalF3100Count  int                               `json:"totalF3100Count"`
	TotalF9100Count  int                               `json:"totalF9100Count"`
}

func CalculateScore(solution map[string]types.LocationSolution, mapName string, generalData types.GeneralGameData, locations []*types.Location) (ScoredSolution, error) {
	scoredSolution := ScoredSolution{
		MapName:   mapName,
		Locations: map[string]types.LocationSolution{},
		GameScore: map[string]float64{
			"co2Savings":    0.0,
			"totalFootfall": 0.0,
		},
	}

	locationListNoRefillStation := map[string]types.Location{}
	for _, loc := range locations {
		key := loc.Name
		newLoc := types.Location{
			Name:              loc.Name,
			Type:              loc.Type,
			Latitude:          loc.Latitude,
			Longitude:         loc.Longitude,
			Footfall:          loc.Footfall,
			FootfallScale:     loc.FootfallScale,
			SalesVolume:       loc.SalesVolume * generalData.RefillSalesFactor,
			NeighborDistances: loc.NeighborDistances,
		}
		if _, ok := solution[key]; !ok {
			locationListNoRefillStation[key] = newLoc
			continue
		}
		f3Count := solution[key].F3
		f9Count := solution[key].F9

		scoredSolution.Locations[key] = types.LocationSolution{
			Location:      newLoc,
			F3:            f3Count,
			F9:            f9Count,
			SalesCapacity: float64(f3Count)*generalData.Freestyle3100Data.RefillCapacityPerWeek + float64(f9Count)*generalData.Freestyle9100Data.RefillCapacityPerWeek,
			LeasingCost:   float64(f3Count)*generalData.Freestyle3100Data.LeasingCostPerWeek + float64(f9Count)*generalData.Freestyle9100Data.LeasingCostPerWeek,
		}
		if scoredSolution.Locations[key].SalesCapacity <= 0 {
			return ScoredSolution{}, fmt.Errorf("location %s has no sales capacity", key)
		}
	}
	if len(scoredSolution.Locations) == 0 {
		return ScoredSolution{}, fmt.Errorf("no locations in solution")
	}

	scoredSolution.Locations = distributeSales(scoredSolution.Locations, locationListNoRefillStation, generalData)
	scoredSolution.Locations = divideFootfall(scoredSolution.Locations, generalData)

	for _, loc := range scoredSolution.Locations {
		newLoc := types.LocationSolution{
			Location: types.Location{
				Name:              loc.Name,
				Type:              loc.Type,
				Latitude:          loc.Latitude,
				Longitude:         loc.Longitude,
				Footfall:          loc.Footfall,
				FootfallScale:     loc.FootfallScale,
				SalesVolume:       math.Round(loc.SalesVolume),
				NeighborDistances: loc.NeighborDistances,
			},
			F9:            loc.F9,
			F3:            loc.F3,
			SalesCapacity: loc.SalesCapacity,
			Revenue:       loc.Revenue,
			Earnings:      loc.Earnings,
			LeasingCost:   loc.LeasingCost,
		}
		sales := newLoc.SalesVolume
		if newLoc.SalesCapacity < newLoc.SalesVolume {
			sales = newLoc.SalesCapacity
		}
		newLoc.Revenue = sales * generalData.RefillUnitData.ProfitPerUnit
		newLoc.Earnings = newLoc.Revenue - newLoc.LeasingCost
		newLoc.Co2Saving = sales*(generalData.ClassicUnitData.Co2PerUnitInGrams-generalData.RefillUnitData.Co2PerUnitInGrams) -
			float64(loc.F3)*generalData.Freestyle3100Data.StaticCo2 -
			float64(loc.F9)*generalData.Freestyle9100Data.StaticCo2

		scoredSolution.Locations[loc.Name] = newLoc

		scoredSolution.TotalF3100Count += loc.F3
		scoredSolution.TotalF9100Count += loc.F9
		scoredSolution.TotalRevenue += newLoc.Revenue
		scoredSolution.TotalLeasingCost += newLoc.LeasingCost
		scoredSolution.GameScore["co2Savings"] += newLoc.Co2Saving / 1000
		scoredSolution.GameScore["totalFootfall"] += loc.Footfall / 1000
	}
	scoredSolution.TotalRevenue = internal.RoundFloatByN(scoredSolution.TotalRevenue, 2)
	scoredSolution.GameScore["co2Savings"] = internal.RoundFloatByN(scoredSolution.GameScore["co2Savings"], 2)
	// scoredSolution.GameScore["co2Savings"] = RoundFloatByN(
	// 	scoredSolution.GameScore["co2Savings"]-
	// 		float64(scoredSolution.TotalF3100Count)*generalData.Freestyle3100Data.StaticCo2/1000-
	// 		float64(scoredSolution.TotalF9100Count)*generalData.Freestyle9100Data.StaticCo2/1000,
	// 	2)
	scoredSolution.GameScore["earnings"] = (scoredSolution.TotalRevenue - scoredSolution.TotalLeasingCost) / 1000
	scoredSolution.GameScore["totalFootfall"] = internal.RoundFloatByN(scoredSolution.GameScore["totalFootfall"], 4)
	scoredSolution.GameScore["total"] = internal.RoundFloatByN(
		(scoredSolution.GameScore["co2Savings"]*generalData.Co2PricePerKiloInSek+scoredSolution.GameScore["earnings"])*
			(1+scoredSolution.GameScore["totalFootfall"]),
		2)
	return scoredSolution, nil
}

func distributeSales(scoredLocations map[string]types.LocationSolution, locationListNoRefillStation map[string]types.Location, generalData types.GeneralGameData) map[string]types.LocationSolution {
	for _, loc := range locationListNoRefillStation {
		key := loc.Name
		distributeTo := make(map[string]float64)
		locationWithoutRefillStation, ok := locationListNoRefillStation[key]
		if !ok {
			continue
		}
		locationsWithinWalkingDistance := loc.NeighborDistances
		total := 0.0

		for locName, dist := range locationsWithinWalkingDistance {
			distributeTo[locName] = math.Pow(generalData.ConstantExpDistributionFunction, generalData.WillingnessToTravelInMeters-dist) - 1.0
			total += distributeTo[locName]
		}

		for locName, dist := range distributeTo {
			if total == 0.0 {
				continue
			}
			newSalesVolume := dist / total * generalData.RefillDistributionRate * locationWithoutRefillStation.SalesVolume
			sLoc := scoredLocations[locName]
			sLoc.SalesVolume += newSalesVolume
			scoredLocations[locName] = sLoc
		}
	}
	return scoredLocations
}

func divideFootfall(scoredLocations map[string]types.LocationSolution, generalData types.GeneralGameData) map[string]types.LocationSolution {
	for key, loc := range scoredLocations {
		count := 1
		for neighbor, distance := range loc.NeighborDistances {
			if key == neighbor {
				continue
			}
			if distance < generalData.WillingnessToTravelInMeters {
				count++
			}
		}
		sLoc := types.LocationSolution{
			Location: types.Location{
				Name:              loc.Name,
				Type:              loc.Type,
				Latitude:          loc.Latitude,
				Longitude:         loc.Longitude,
				Footfall:          loc.Footfall / float64(count),
				FootfallScale:     loc.FootfallScale,
				SalesVolume:       loc.SalesVolume,
				NeighborDistances: loc.NeighborDistances,
			},
			F9:            loc.F9,
			F3:            loc.F3,
			SalesCapacity: loc.SalesCapacity,
			Revenue:       loc.Revenue,
			Earnings:      loc.Earnings,
			LeasingCost:   loc.LeasingCost,
		}
		scoredLocations[key] = sLoc
	}
	return scoredLocations
}
