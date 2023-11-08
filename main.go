package main

import "fmt"

func main() {

	const APIKey = "74266cdf-1f38-403c-8766-044cc03d9162"
	const BaseURL = "https://api.considition.com"
	client := NewClient(APIKey, BaseURL)
	mapData, err := client.GetMapData("uppsala")
	if err != nil {
		panic(err)
	}
	generalGameData, err := client.GetGeneralData()
	if err != nil {
		panic(err)
	}

	solution := make(map[string]LocationSolution)

	for _, loc := range mapData.Locations {
		f9 := 0
		if loc.Name == "location55" {
			f9 = 1
		}
		solution[loc.Name] = LocationSolution{
			Location:      loc,
			F3:            0,
			F9:            f9,
			SalesCapacity: 0,
			Revenue:       0,
			Earnings:      0,
			LeasingCost:   0,
		}
	}
	filterEmptyLocations := func(solution map[string]LocationSolution) map[string]LocationSolution {
		filtered := make(map[string]LocationSolution)
		for _, loc := range solution {
			if loc.F3 > 0 || loc.F9 > 0 {
				filtered[loc.Location.Name] = loc
			}
		}
		return filtered
	}
	filtered := filterEmptyLocations(solution)

	scoredSolution, err := CalculateScore("uppsala", filtered, *mapData, *generalGameData)
	if err != nil {
		panic(err)
	}
	fmt.Printf("%v\n", scoredSolution)

	// retSol, err := client.SubmitSolution("uppsala", scoredSolution)
	// if err != nil {
	// 	panic(err)
	// }
	// fmt.Printf("%v\n", retSol)
	// fmt.Println(scoredSolution.GameScore["total"])

}
