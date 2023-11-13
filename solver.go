package main

import (
	"encoding/csv"
	"fmt"
	"log"
	"math/rand"
	"os"
	"sort"
	"strconv"
	"sync"
	"time"
)

type SolverConfig struct {
	GenerationLimit            int
	PopulationSize             int
	Locations                  []*Location
	MapData                    *MapData
	GeneralGameData            *GeneralGameData
	MutationProbability        float64
	ReproductionMethod         ReproductionMethod
	GenerationImprovementLimit int
}

type ReproductionMethod string

var (
	RouletteWheel               ReproductionMethod = "rouletteWheel"
	Tournament                  ReproductionMethod = "tournament"
	RankSelection               ReproductionMethod = "rankSelection"
	StochasticUniversalSampling ReproductionMethod = "stochasticUniversalSampling"
)

type Solver struct {
	Config            SolverConfig
	Population        []*Genome
	BestGenome        *Genome
	BestSolution      float64
	WorstSolution     float64
	AverageScore      float64
	RNG               *rand.Rand
	OptLog            []OptimizationLog
	LatestImprovement int
	Generation        int
}

type OptimizationLog struct {
	Generation    int     `csv:"generation"`
	BestSolution  float64 `csv:"best_solution"`
	WorstSolution float64 `csv:"worst_solution"`
	AverageScore  float64 `csv:"average_score"`
}

func NewSolver(cfg SolverConfig) *Solver {
	return &Solver{
		Config: cfg,
	}
}

func (s *Solver) Run() {
	s.SeedPopulation()
	for generation := 0; generation-s.LatestImprovement < s.Config.GenerationImprovementLimit; generation++ {
		s.Generation = generation
		// fmt.Printf("Generation %d:\t", generation)
		s.EvaluatePopulation()
		s.RankPopulation(generation)
		// fmt.Printf("Best solution: %f\t", s.BestSolution)
		// fmt.Printf("Worst solution: %f\t", s.WorstSolution)
		// fmt.Printf("Average score: %f\n", s.AverageScore)
		s.OptLog = append(s.OptLog, OptimizationLog{
			Generation:    generation,
			BestSolution:  s.BestSolution,
			WorstSolution: s.WorstSolution,
			AverageScore:  s.AverageScore,
		})
		cloningSelection := s.SelectForCloning()
		cloned := s.Clone(cloningSelection)
		crossoverSelection := s.SelectForCrossover()
		babies := s.Crossover(crossoverSelection)
		newRandomGenomes := make([]*Genome, 10)
		for i := 0; i < 10; i++ {
			newRandomGenomes[i] = NewRandomGenome(s.RNG, len(s.Config.Locations))
		}
		s.Replace(cloned, babies, newRandomGenomes)
		s.Mutate()
	}
}

func (s *Solver) SeedPopulation() {
	s.RNG = rand.New(rand.NewSource(time.Now().UnixNano()))
	for i := 0; i < s.Config.PopulationSize; i++ {
		s.Population = append(s.Population, NewRandomGenome(s.RNG, len(s.Config.Locations)))
	}
}

func (s *Solver) EvaluatePopulation() {
	var wg sync.WaitGroup

	for _, genome := range s.Population {
		wg.Add(1) // Increment the WaitGroup counter
		go func(g *Genome) {
			defer wg.Done() // Decrement the counter when the goroutine completes
			g.Evaluate(s.Config.Locations, *s.Config.MapData, *s.Config.GeneralGameData)
		}(genome)
	}

	wg.Wait() // Wait for all goroutines to finish
}

func (s *Solver) RankPopulation(generation int) {
	sort.Slice(s.Population, func(i, j int) bool {
		return s.Population[i].Score > s.Population[j].Score
	})
	s.BestSolution = s.Population[0].Score
	if s.BestGenome == nil || s.BestGenome.Score < s.BestSolution {
		s.BestGenome = s.Population[0].Copy()
		s.LatestImprovement = generation
		fmt.Printf("(Generation %d)\tNew best: %.f\n", s.Generation, s.BestGenome.Score)
	}
	s.Population = append([]*Genome{s.BestGenome.Copy()}, s.Population[:s.Config.PopulationSize-1]...)
	s.WorstSolution = s.Population[len(s.Population)-1].Score
	s.AverageScore = 0
	for _, genome := range s.Population {
		s.AverageScore += genome.Score
	}
	s.AverageScore /= float64(len(s.Population))
}

func (s *Solver) SelectForCrossover() []*Genome {
	selection := make([]*Genome, len(s.Population)-len(s.Population)/10-10)
	for i := 0; i < len(s.Population)-len(s.Population)/10-10; i++ {
		selection[i] = s.Population[i]
	}
	return selection
}

func (s *Solver) SelectForCloning() []*Genome {
	selection := make([]*Genome, len(s.Population)/10)
	for i := 0; i < len(s.Population)/10; i++ {
		selection[i] = s.Population[i]
	}
	return selection
}

func (s *Solver) SelectForReplacement() []*Genome {
	popLen := len(s.Population) / 2
	toBeReplaced := make([]*Genome, popLen)
	for i := popLen; i < len(s.Population); i++ {
		toBeReplaced[i-popLen] = s.Population[i]
	}
	return toBeReplaced
}

func (s *Solver) Mutate() {
	for i, genome := range s.Population {
		if i == 0 {
			continue // don't mutate the best genome
		}
		// if mutate threshold is met, mutate genome
		// if rand.Float64() < s.Config.MutationProbability {
		if rand.Float64() < genome.Score/s.BestGenome.Score {
			genome.Mutate2(s.Config.MutationProbability)
		}
	}
}

func (s *Solver) Crossover(population []*Genome) []*Genome {
	randomOrder := s.RNG.Perm(len(population))
	babies := make([]*Genome, len(population))
	for i := 0; i < len(population); i += 2 {
		babies[i], babies[i+1] = population[randomOrder[i]].Crossover(population[randomOrder[i+1]])
		// babies[i], babies[i+1] = population[randomOrder[i]].CrossoverSinglePair(population[randomOrder[i+1]])
	}
	return babies
}

func (s *Solver) Clone(genomes []*Genome) []*Genome {
	clones := make([]*Genome, len(genomes))
	for i := 0; i < len(genomes); i++ {
		clones[i] = genomes[i].Copy()
	}
	return clones
}

func (s *Solver) Replace(cloned, babies, randomGenes []*Genome) {
	newPopulation := make([]*Genome, len(s.Population))
	for i := 0; i < len(cloned); i++ {
		newPopulation[i] = cloned[i]
	}
	for i := 0; i < len(babies); i++ {
		newPopulation[i+len(cloned)] = babies[i]
	}
	for i := 0; i < len(randomGenes); i++ {
		newPopulation[i+len(cloned)+len(babies)] = randomGenes[i]
	}
	s.Population = newPopulation
}

func (s *Solver) GetSolution() ScoredSolution {
	genome := s.BestGenome
	genomeLocation := make(map[string]LocationSolution)
	for j, loc := range s.Config.Locations {
		genomeLocation[loc.Name] = LocationSolution{
			Location: *loc,
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
	scoredSolution, err := CalculateScore(filtered, s.Config.MapData.Name, *s.Config.GeneralGameData, s.Config.Locations)
	if err != nil {
		panic(err)
	}
	return scoredSolution
}

func (s *Solver) WriteOptimizationLogToFile(file *os.File) error {
	// Create a CSV writer
	writer := csv.NewWriter(file)
	defer writer.Flush()

	// Write the CSV header
	header := []string{"generation", "best_solution", "worst_solution", "average_score"}
	if err := writer.Write(header); err != nil {
		log.Fatal(err)
	}

	// Write each struct as a row in the CSV file
	for _, logEntry := range s.OptLog {
		record := []string{
			strconv.Itoa(logEntry.Generation),
			strconv.FormatFloat(logEntry.BestSolution, 'f', -1, 64),
			strconv.FormatFloat(logEntry.WorstSolution, 'f', -1, 64),
			strconv.FormatFloat(logEntry.AverageScore, 'f', -1, 64),
		}

		if err := writer.Write(record); err != nil {
			log.Fatal(err)
		}
	}
	return nil
}
