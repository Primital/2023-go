package main

import (
	"fmt"
	"os"
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
		GenerationLimit:     1000,
		PopulationSize:      1000,
		Locations:           locations,
		MapData:             mapData,
		GeneralGameData:     generalGameData,
		MutationProbability: 0.1,
	}

	// Profile
	// f, err := os.Create("cpu.prof")
	// if err != nil {
	// 	log.Fatal("could not create CPU profile: ", err)
	// }
	// defer f.Close()
	//
	// if err := pprof.StartCPUProfile(f); err != nil {
	// 	log.Fatal("could not start CPU profile: ", err)
	// }
	// defer pprof.StopCPUProfile()

	solver := NewSolver(solverConfig)
	solver.Run()

	// write csv of optimization log to file
	f, err := os.Create("optlog.csv")
	if err != nil {
		panic(err)
	}
	defer f.Close()
	solver.WriteOptimizationLogToFile(f)

	if submit {
		responseSol, err := client.SubmitSolution(mapData.Name, solver.GetSolution())
		if err != nil {
			panic(err)
		}
		fmt.Printf("Submitted game %s\n", responseSol.ID)
		fmt.Printf("Response: %v", responseSol.Score.Total)
	}
}
