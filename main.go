package main

import (
	"fmt"
)

const submit = false

func main() {
	const APIKey = "74266cdf-1f38-403c-8766-044cc03d9162"
	const BaseURL = "https://api.considition.com"
	client := NewClient(APIKey, BaseURL)
	mapData, err := client.GetMapData("goteborg")
	if err != nil {
		panic(err)
	}
	generalGameData, err := client.GetGeneralData()
	if err != nil {
		panic(err)
	}

	locations := make([]*Location, len(mapData.Locations))
	for i := 1; i <= len(mapData.Locations); i++ {
		mapLoc := mapData.Locations[fmt.Sprintf("location%d", i)]
		locations[i-1] = &mapLoc
	}
	PrecalculateNeighborDistances(locations)
	solverConfig := SolverConfig{
		GenerationLimit:     2000,
		PopulationSize:      300,
		Locations:           locations,
		MapData:             mapData,
		GeneralGameData:     generalGameData,
		MutationProbability: 0.1,
	}

	solver := NewSolver(solverConfig)
	solver.Run()
	if submit {
		responseSol, err := client.SubmitSolution(mapData.Name, solver.GetSolution())
		if err != nil {
			panic(err)
		}
		fmt.Printf("Submitted game %s\n", responseSol.ID)
		fmt.Printf("Response: %v", responseSol.Score.Total)
	}
}
