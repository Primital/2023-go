package main

import (
	"fmt"
	"math"

	"github.com/google/uuid"
)

type LocationSolution struct {
	Location
	F9            int     `json:"f9"`
	F3            int     `json:"f3"`
	SalesCapacity float64 `json:"salesCapacity"`
	Revenue       float64 `json:"revenue"`
	Earnings      float64 `json:"earnings"`
	LeasingCost   float64 `json:"leasingCost"`
}

type ScoredSolution struct {
	GameId           string                      `json:"gameId"`
	MapName          string                      `json:"mapName"`
	Locations        map[string]LocationSolution `json:"locations"`
	GameScore        map[string]float64          `json:"gameScore"`
	TotalRevenue     float64                     `json:"totalRevenue"`
	TotalLeasingCost float64                     `json:"totalLeasingCost"`
	TotalF3100Count  int                         `json:"totalF3100Count"`
	TotalF9100Count  int                         `json:"totalF9100Count"`
}

func CalculateScore(mapName string, solution map[string]LocationSolution, mapData MapData, generalData GeneralGameData) (ScoredSolution, error) {
	scoredSolution := ScoredSolution{
		GameId:    uuid.New().String(),
		MapName:   mapName,
		Locations: map[string]LocationSolution{},
		GameScore: map[string]float64{
			"co2Savings":    0.0,
			"totalFootfall": 0.0,
		},
	}

	locationListNoRefillStation := map[string]Location{}
	for _, loc := range mapData.Locations {
		key := loc.Name
		if _, ok := solution[key]; ok {
			f3Count := solution[key].F3
			f9Count := solution[key].F9

			newLoc := Location{
				Name:          loc.Name,
				Type:          loc.Type,
				Latitude:      loc.Latitude,
				Longitude:     loc.Longitude,
				Footfall:      loc.Footfall,
				FootfallScale: loc.FootfallScale,
				SalesVolume:   loc.SalesVolume * generalData.RefillSalesFactor,
			}
			scoredSolution.Locations[key] = LocationSolution{
				Location:      newLoc,
				F3:            f3Count,
				F9:            f9Count,
				SalesCapacity: float64(f3Count)*generalData.Freestyle3100Data.RefillCapacityPerWeek + float64(f9Count)*generalData.Freestyle9100Data.RefillCapacityPerWeek,
				LeasingCost:   float64(f3Count)*generalData.Freestyle3100Data.LeasingCostPerWeek + float64(f9Count)*generalData.Freestyle9100Data.LeasingCostPerWeek,
			}
			if scoredSolution.Locations[key].SalesCapacity <= 0 {
				return ScoredSolution{}, fmt.Errorf("location %s has no sales capacity", key)
			}
		} else {
			newLoc := loc
			newLoc.SalesVolume = loc.SalesVolume * generalData.RefillSalesFactor
			locationListNoRefillStation[key] = newLoc
		}
	}
	if len(scoredSolution.Locations) == 0 {
		return ScoredSolution{}, fmt.Errorf("no locations in solution")
	}

	scoredSolution.Locations = distributeSales(scoredSolution.Locations, locationListNoRefillStation, generalData)

	for _, loc := range scoredSolution.Locations {
		newLoc := LocationSolution{
			Location: Location{
				Name:          loc.Name,
				Type:          loc.Type,
				Latitude:      loc.Latitude,
				Longitude:     loc.Longitude,
				Footfall:      loc.Footfall,
				FootfallScale: loc.FootfallScale,
				SalesVolume:   loc.SalesVolume,
			},
			F9:            loc.F9,
			F3:            loc.F3,
			SalesCapacity: loc.SalesCapacity,
			Revenue:       loc.Revenue,
			Earnings:      loc.Earnings,
			LeasingCost:   loc.LeasingCost,
		}
		newLoc.SalesVolume = math.Round(loc.SalesVolume)
		sales := newLoc.SalesVolume
		if newLoc.SalesCapacity < newLoc.SalesVolume {
			sales = newLoc.SalesCapacity
		}
		newLoc.Revenue = sales * generalData.RefillUnitData.ProfitPerUnit
		newLoc.Earnings = newLoc.Revenue - newLoc.LeasingCost

		scoredSolution.Locations[loc.Name] = newLoc

		scoredSolution.TotalF3100Count += loc.F3
		scoredSolution.TotalF9100Count += loc.F9
		scoredSolution.TotalRevenue += newLoc.Revenue
		scoredSolution.TotalLeasingCost += newLoc.LeasingCost
		scoredSolution.GameScore["co2Savings"] += sales * (generalData.ClassicUnitData.Co2PerUnitInGrams - generalData.RefillUnitData.Co2PerUnitInGrams) / 1000
		scoredSolution.GameScore["totalFootfall"] += loc.Footfall
	}
	scoredSolution.TotalRevenue = math.Round(scoredSolution.TotalRevenue)
	scoredSolution.GameScore["co2Savings"] = math.Round(scoredSolution.GameScore["co2Savings"] -
		float64(scoredSolution.TotalF3100Count)*generalData.Freestyle3100Data.StaticCo2/1000 -
		float64(scoredSolution.TotalF9100Count)*generalData.Freestyle9100Data.StaticCo2/1000)
	scoredSolution.GameScore["earnings"] = scoredSolution.TotalRevenue - scoredSolution.TotalLeasingCost
	scoredSolution.GameScore["total"] = math.Round((scoredSolution.GameScore["co2Savings"]*generalData.Co2PricePerKiloInSek + scoredSolution.GameScore["earnings"]) * (1 + scoredSolution.GameScore["totalFootfall"]))
	return scoredSolution, nil
}

func distributeSales(scoredLocations map[string]LocationSolution, locationListNoRefillStation map[string]Location, generalData GeneralGameData) map[string]LocationSolution {
	for _, loc := range locationListNoRefillStation {
		key := loc.Name
		distributeTo := make(map[string]float64)
		locationWithoutRefillStation, ok := locationListNoRefillStation[key]
		if !ok {
			continue
		}
		// Just put this in a global precalculated map if this is slow
		locationsWithInVicinity := loc.GetLocationsWithinWalkingDistance(scoredLocations, generalData)
		total := 0.0

		for locName, dist := range locationsWithInVicinity {
			distributeTo[locName] = math.Pow(generalData.ConstantExpDistributionFunction, generalData.WillingnessToTravelInMeters-dist) - 1.0
			total += distributeTo[locName]
		}

		for locName, dist := range distributeTo {
			newSalesVolume := dist / total * generalData.RefillDistributionRate * locationWithoutRefillStation.SalesVolume
			sLoc := scoredLocations[locName]
			sLoc.SalesVolume += newSalesVolume
			scoredLocations[locName] = sLoc
		}
	}
	return scoredLocations
}
