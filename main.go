package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"2023-go/api"
	"2023-go/internal"
	solver2 "2023-go/solver"
	"2023-go/types"
)

const submit = true
const debug = false
const logOptimization = false

func main() {
	seedFilePath := ""
	if len(os.Args) > 1 {
		seedFilePath = os.Args[1]
		fmt.Printf("Using seed file %s\n", seedFilePath)
	}

	// Create a channel to receive signals.
	sigCh := make(chan os.Signal, 1)

	// Register the SIGINT signal (interrupt signal) to the signal channel.
	signal.Notify(sigCh, syscall.SIGINT)

	const APIKey = "74266cdf-1f38-403c-8766-044cc03d9162"
	const BaseURL = "https://api.considition.com"
	client := api.NewClient(APIKey, BaseURL)
	mapData, err := client.GetMapData("berlin")
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
	solverConfig := solver2.Config{
		PopulationSize:             200,
		Locations:                  locations,
		MapData:                    mapData,
		GeneralGameData:            generalGameData,
		MutationProbability:        0.4,
		GenerationImprovementLimit: 2000,
	}

	solver := solver2.NewSolver(solverConfig)

	// Goroutine to handle signals.
	go func(solver *solver2.Solver) {
		for {
			select {
			case sig := <-sigCh:
				switch sig {
				case syscall.SIGINT:
					fmt.Println("Received SIGINT. Submitting and saving")
					solver.Finish()
				}
			}
		}
	}(solver)

	if seedFilePath != "" {
		fmt.Printf("Loading seed file %s\n", seedFilePath)
		seedFile, err := os.Open(seedFilePath)
		if err != nil {
			panic(err)
		}
		if err := solver.LoadSolutionFromFile(seedFile); err != nil {
			panic(err)
		}
		seedFile.Close()
	}

	solver.Optimize(debug)

	fmt.Printf("Best solution (Generation %d): %.f\n", solver.LatestImprovement, solver.BestGenome.Score)

	// // write csv of optimization log to file
	if logOptimization {
		now := time.Now()
		optLogFile, err := os.Create(fmt.Sprintf("optlogs/optlog-%s-%s.csv", mapData.Name, now.Format("2006-01-02-15-04-05")))
		if err != nil {
			panic(err)
		}
		defer optLogFile.Close()
		solver.WriteOptimizationLogToFile(optLogFile)
	}

	if submit {
		sol := solver.GetSolution()
		responseSol, err := client.SubmitSolution(mapData.Name, sol)
		if err != nil {
			panic(err)
		}
		// fmt.Printf("Submitted game %s\n", responseSol.ID)
		fmt.Printf("Response: %v\n", responseSol.Score.Total)

		// fmt.Printf("Writing solution to file\n")
		if err := solver.WriteSolutionToFile(responseSol.ID); err != nil {
			panic(err)
		}
	}
}
