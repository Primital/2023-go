package main

import (
	"fmt"
	"os"
	"time"

	"2023-go/internal"
	"2023-go/types"
)

const submit = true

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

	locations := make([]*types.Location, len(mapData.Locations))
	for i := 1; i <= len(mapData.Locations); i++ {
		mapLoc := mapData.Locations[fmt.Sprintf("location%d", i)]
		locations[i-1] = &mapLoc
	}
	internal.PrecalculateNeighborDistances(locations, generalGameData)
	solverConfig := SolverConfig{
		PopulationSize:             500,
		Locations:                  locations,
		MapData:                    mapData,
		GeneralGameData:            generalGameData,
		MutationProbability:        0.4,
		GenerationImprovementLimit: 10000,
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
	solver.Optimize()

	fmt.Printf("Best solution (Generation %d): %.f\n", solver.LatestImprovement, solver.BestGenome.Score)

	// write csv of optimization log to file
	now := time.Now()
	f, err := os.Create(fmt.Sprintf("optlogs/optlog-%s-%s.csv", mapData.Name, now.Format("2006-01-02-15-04-05")))
	if err != nil {
		panic(err)
	}
	defer f.Close()
	solver.WriteOptimizationLogToFile(f)

	if submit {
		sol := solver.GetSolution()
		responseSol, err := client.SubmitSolution(mapData.Name, sol)
		if err != nil {
			panic(err)
		}
		fmt.Printf("Submitted game %s\n", responseSol.ID)
		fmt.Printf("Response: %v", responseSol.Score.Total)
		solver.GetSolution()
	}
}
