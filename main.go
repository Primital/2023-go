package main

import (
	"fmt"
	"os"
	"time"

	"2023-go/api"
	"2023-go/internal"
	solver2 "2023-go/solver"
	"2023-go/types"
)

const submit = false
const debug = true

func main() {
	const APIKey = "74266cdf-1f38-403c-8766-044cc03d9162"
	const BaseURL = "https://api.considition.com"
	client := api.NewClient(APIKey, BaseURL)
	mapData, err := client.GetMapData("goteborg")
	if err != nil {
		panic(err)
	}
	generalGameData, err := client.GetGeneralData()
	if err != nil {
		panic(err)
	}

	locations := make([]*types.Location, len(mapData.Locations))
	for i := 0; i < len(mapData.Locations); i++ {
		mapLoc := mapData.Locations[fmt.Sprintf("location%d", i+1)]
		locations[i] = &mapLoc
	}
	internal.PrecalculateNeighborDistances(locations, generalGameData)
	solverConfig := solver2.SolverConfig{
		PopulationSize:             500,
		Locations:                  locations,
		MapData:                    mapData,
		GeneralGameData:            generalGameData,
		MutationProbability:        0.3,
		GenerationImprovementLimit: 5000,
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

	solver := solver2.NewSolver(solverConfig)
	solver.Optimize(debug)

	fmt.Printf("Best solution (Generation %d): %.f\n", solver.LatestImprovement, solver.BestGenome.Score)
	//
	// // write csv of optimization log to file
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
		// fmt.Printf("Submitted game %s\n", responseSol.ID)
		fmt.Printf("Response: %v\n", responseSol.Score.Total)

		// fmt.Printf("Writing solution to file\n")
		f, err := os.Create(fmt.Sprintf("solutions/solution-%s-%s.csv", mapData.Name, now.Format("2006-01-02-15-04-05")))
		if err != nil {
			panic(err)
		}
		defer f.Close()
		if err := solver.WriteSolutionToFile(f, responseSol.ID); err != nil {
			panic(err)
		}
	}
}
