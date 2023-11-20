package scoring_test

import (
	"encoding/json"
	"fmt"
	"os"
	"testing"

	"2023-go/genome"
	"2023-go/scoring"
	"2023-go/types"
)

func Test_ScoreSolution(t *testing.T) {
	validationSolution, err := LoadValidationData()
	if err != nil {
		panic(err)
	}
	fmt.Println(validationSolution)
	// Load map locations
	f, err := os.OpenFile("../map_data/uppsala.json", os.O_RDONLY, 0644)
	if err != nil {
		panic(err)
	}

	var mapData types.MapData
	if err := json.NewDecoder(f).Decode(&mapData); err != nil {
		panic(err)
	}
	f.Close()

	// Load general game data
	f, err = os.OpenFile("general_data.json", os.O_RDONLY, 0644)
	if err != nil {
		panic(err)
	}
	defer f.Close()

	var generalGameData types.GeneralGameData
	if err := json.NewDecoder(f).Decode(&generalGameData); err != nil {
		panic(err)
	}

	orderedLocations := make([]*types.Location, len(mapData.Locations))
	pairs := make([]genome.Pair, len(mapData.Locations))
	for i := 1; i <= len(mapData.Locations); i++ {
		locName := fmt.Sprintf("location%d", i)
		loc := mapData.Locations[locName]
		orderedLocations[i-1] = &loc
		validationLocation, ok := validationSolution.Locations[locName]
		if ok {
			pairs[i-1] = genome.Pair{
				F3: validationLocation.Freestyle3100Count,
				F9: validationLocation.Freestyle9100Count,
			}
		}
	}

	// Populate genome with values from validationSolution
	bestGenome := genome.Genome{
		Pairs: pairs,
	}
	genomeLocation := make(map[string]types.LocationSolution)
	for j, loc := range orderedLocations {
		genomeLocation[loc.Name] = types.LocationSolution{
			Location: *loc,
			F3:       bestGenome.Pairs[j].F3,
			F9:       bestGenome.Pairs[j].F9,
		}
	}

	filterEmptyLocations := func(solution map[string]types.LocationSolution) map[string]types.LocationSolution {
		filtered := make(map[string]types.LocationSolution)
		for _, loc := range solution {
			if loc.F3 > 0 || loc.F9 > 0 {
				filtered[loc.Location.Name] = loc
			}
		}
		return filtered
	}
	filtered := filterEmptyLocations(genomeLocation)
	scoredSolution, err := scoring.CalculateScore(filtered, mapData.Name, generalGameData, orderedLocations)
	if err != nil {
		panic(err)
	}

	// Compare scoredSolution with validationSolution
	if got, want := scoredSolution.GameScore.KgCo2Savings, validationSolution.GameScore.KgCo2Savings; got != want {
		t.Errorf("KgCo2Savings: got %f, want %f", got, want)
	}
	if got, want := scoredSolution.GameScore.TotalFootfall, validationSolution.GameScore.TotalFootfall; got != want {
		t.Errorf("TotalFootfall: got %f, want %f", got, want)
	}
	if got, want := scoredSolution.GameScore.Earnings, validationSolution.GameScore.Earnings; got != want {
		t.Errorf("Earnings: got %f, want %f", got, want)
	}
	if got, want := scoredSolution.GameScore.Total, validationSolution.GameScore.Total; got != want {
		t.Errorf("Total: got %f, want %f", got, want)
	}
	if got, want := scoredSolution.TotalRevenue, validationSolution.TotalRevenue; got != want {
		t.Errorf("TotalRevenue: got %f, want %f", got, want)
	}
	if got, want := scoredSolution.TotalLeasingCost, validationSolution.TotalLeasingCost; got != want {
		t.Errorf("TotalLeasingCost: got %f, want %f", got, want)
	}
	if got, want := scoredSolution.TotalF3100Count, validationSolution.TotalF3100Count; got != want {
		t.Errorf("TotalF3100Count: got %d, want %d", got, want)
	}
	if got, want := scoredSolution.TotalF9100Count, validationSolution.TotalF9100Count; got != want {
		t.Errorf("TotalF9100Count: got %d, want %d", got, want)
	}
	for _, loc := range orderedLocations {
		validationLocation, ok := validationSolution.Locations[loc.Name]
		if !ok {
			continue
		}
		scoredLocation, ok := scoredSolution.Locations[loc.Name]
		if !ok {
			continue
		}
		if got, want := scoredLocation.Location.Name, validationLocation.LocationName; got != want {
			t.Errorf("LocationName: got %s, want %s", got, want)
		}
		if got, want := scoredLocation.LeasingCost, validationLocation.LeasingCost; got != want {
			t.Errorf("%s, LeasingCost: got %f, want %f", loc.Name, got, want)
		}
		if got, want := scoredLocation.Revenue, validationLocation.Revenue; got != want {
			t.Errorf("%s, Revenue: got %f, want %f", loc.Name, got, want)
		}
		if got, want := scoredLocation.Earnings, validationLocation.Earnings; got != want {
			t.Errorf("%s, Earnings: got %f, want %f", loc.Name, got, want)
		}
		if got, want := scoredLocation.F9, validationLocation.Freestyle9100Count; got != want {
			t.Errorf("%s, Freestyle9100Count: got %d, want %d", loc.Name, got, want)
		}
		if got, want := scoredLocation.F3, validationLocation.Freestyle3100Count; got != want {
			t.Errorf("%s, Freestyle3100Count: got %d, want %d", loc.Name, got, want)
		}
		if got, want := scoredLocation.Co2Saving, validationLocation.GramCo2Savings; got != want {
			t.Errorf("%s, Co2Saving: got %f, want %f", loc.Name, got, want)
		}
		if got, want := scoredLocation.SalesCapacity, validationLocation.SalesCapacity; got != want {
			t.Errorf("%s, SalesCapacity: got %f, want %f", loc.Name, got, want)
		}
		if got, want := scoredLocation.SalesVolume, validationLocation.SalesVolume; got != want {
			t.Errorf("%s, SalesVolume: got %f, want %f", loc.Name, got, want)
		}
		if got, want := scoredLocation.Footfall, validationLocation.Footfall; got != want {
			t.Errorf("%s, Footfall: got %f, want %f", loc.Name, got, want)
		}
		if got, want := scoredLocation.FootfallScale, validationLocation.FootfallScale; got != want {
			t.Errorf("%s, FootfallScale: got %f, want %f", loc.Name, got, want)
		}

	}

}

type (
	TestScore struct {
		GameScore        TestGameScore            `json:"gameScore"`
		Locations        map[string]LocationScore `json:"locations"`
		TotalRevenue     float64                  `json:"totalRevenue"`
		TotalLeasingCost float64                  `json:"totalLeasingCost"`
		TotalF3100Count  int                      `json:"totalFreestyle3100Count"`
		TotalF9100Count  int                      `json:"totalFreestyle9100Count"`
	}

	TestGameScore struct {
		KgCo2Savings  float64 `json:"kgCo2Savings"`
		TotalFootfall float64 `json:"totalFootfall"`
		Earnings      float64 `json:"earnings"`
		Total         float64 `json:"total"`
	}

	LocationScore struct {
		LocationName       string  `json:"locationName"`
		LocationType       string  `json:"locationType"`
		Latitude           float64 `json:"latitude"`
		Longitude          float64 `json:"longitude"`
		Footfall           float64 `json:"footfall"`
		FootfallScale      float64 `json:"footfallScale"`
		SalesVolume        float64 `json:"salesVolume"`
		SalesCapacity      float64 `json:"salesCapacity"`
		LeasingCost        float64 `json:"leasingCost"`
		Revenue            float64 `json:"revenue"`
		Earnings           float64 `json:"earnings"`
		Freestyle9100Count int     `json:"freestyle9100Count"`
		Freestyle3100Count int     `json:"freestyle3100Count"`
		GramCo2Savings     float64 `json:"gramCo2Savings"`
		IsProfitable       bool    `json:"isProfitable"`
		IsCo2Saving        bool    `json:"isCo2Saving"`
	}
)

func LoadValidationData() (TestScore, error) {
	f, err := os.OpenFile("validation_uppsala.json", os.O_RDONLY, 0644)
	if err != nil {
		return TestScore{}, err
	}
	defer f.Close()

	var testScore TestScore
	if err := json.NewDecoder(f).Decode(&testScore); err != nil {
		return TestScore{}, err
	}
	return testScore, nil
}
