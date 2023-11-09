package main

import (
	"fmt"
	"math/rand"
	"sort"
	"time"
)

/*
IDEA:
- Create a solver that takes a map and a general game data as input
- The solver should instantiate a population of solutions
- Each solution is evaulated and ranked (sorted) based on its score
- The top X% of solutions are selected for reproduction
- The top Y% of solutions are selected for mutation
- The top Z% of solutions are selected for crossover
- The top A% of solutions are selected for cloning
- The bottom B% of solutions are selected for replacement
- The solver keeps track of the best solutions, and the worst
- The solver keeps track of the average score of the population
- Fiddle with how high the chance for mutation and stuff should be
- If possible, make the solver run in parallel
- If possible, make the solver draw a graph of how the solution is improving over time


TODO: Implement function for seeding the population with random solutions
TODO: Implement function for mutating solution
TODO: Implement function for crossover
TODO: Implement function for cloning
TODO: Implement function for replacing
DONE: Implement function for evaluating a solution
*/

type SolverConfig struct {
	GenerationLimit     int
	PopulationSize      int
	Locations           []Location
	MapData             *MapData
	GeneralGameData     *GeneralGameData
	MutationProbability float64
}

type Solver struct {
	Config        SolverConfig
	Population    []*Genome
	BestSolution  float64
	WorstSolution float64
	AverageScore  float64
	RNG           *rand.Rand
}

func NewSolver(cfg SolverConfig) *Solver {
	return &Solver{
		Config: cfg,
	}
}

func (s *Solver) Run() {
	s.SeedPopulation()
	for generation := 0; generation < s.Config.GenerationLimit; generation++ {
		fmt.Printf("Generation %d:\t", generation)
		s.EvaluatePopulation()
		s.RankPopulation()
		fmt.Printf("Best solution: %f\t", s.BestSolution)
		fmt.Printf("Worst solution: %f\t", s.WorstSolution)
		fmt.Printf("Average score: %f\n", s.AverageScore)
		s.Mutate()
		s.Crossover()
		s.Clone()
		s.Replace()
	}
}

func (s *Solver) SeedPopulation() {
	s.RNG = rand.New(rand.NewSource(time.Now().UnixNano()))
	for i := 0; i < s.Config.PopulationSize; i++ {
		s.Population = append(s.Population, NewRandomGenome(s.RNG, len(s.Config.Locations)))
	}
}

func (s *Solver) EvaluatePopulation() {
	for _, genome := range s.Population {
		genomeLocation := make(map[string]LocationSolution)
		for j, loc := range s.Config.Locations {
			genomeLocation[loc.Name] = LocationSolution{
				Location: loc,
				F3:       genome.Pairs[j].F3,
				F9:       genome.Pairs[j].F9,
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
		filtered := filterEmptyLocations(genomeLocation)
		if len(filtered) == 0 {
			genome.Score = 0
			continue
		}
		scoredSolution, err := CalculateScore(filtered, *s.Config.MapData, *s.Config.GeneralGameData)
		if err != nil {
			panic(err)
		}
		genome.Score = scoredSolution.GameScore["total"]
	}
}

func (s *Solver) RankPopulation() {
	sort.Slice(s.Population, func(i, j int) bool {
		return s.Population[i].Score > s.Population[j].Score
	})
	s.BestSolution = s.Population[0].Score
	s.WorstSolution = s.Population[len(s.Population)-1].Score
	s.AverageScore = 0
	for _, genome := range s.Population {
		s.AverageScore += genome.Score
	}
	s.AverageScore /= float64(len(s.Population))
}

func (s *Solver) Mutate() {
	for i, genome := range s.Population {
		if i == 0 {
			// Keep the best
			continue
		}
		genome.Mutate(s.Config.MutationProbability)
	}
}

func (s *Solver) Crossover() {
	// TODO: Implement
}

func (s *Solver) Clone() {
	// TODO: Implement
}

func (s *Solver) Replace() {
	// TODO: Implement
}
