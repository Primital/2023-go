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

	locations := make([]Location, len(mapData.Locations))
	for i := 1; i <= len(mapData.Locations); i++ {
		locations[i-1] = mapData.Locations[fmt.Sprintf("location%d", i)]
	}
	solverConfig := SolverConfig{
		GenerationLimit:     50,
		PopulationSize:      1000,
		Locations:           locations,
		MapData:             mapData,
		GeneralGameData:     generalGameData,
		MutationProbability: 0.1,
	}

	solver := NewSolver(solverConfig)
	solver.Run()
}
