package solver

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"log"
	"math"
	"math/rand"
	"os"
	"sort"
	"strconv"
	"sync"
	"time"

	"2023-go/genome"
	"2023-go/scoring"
	"2023-go/types"
)

type Config struct {
	GenerationLimit            int
	PopulationSize             int
	Locations                  []*types.Location
	MapData                    *types.MapData
	GeneralGameData            *types.GeneralGameData
	MutationProbability        float64
	ReproductionMethod         ReproductionMethod // Intended to be used, but just used RankSelection
	GenerationImprovementLimit int
	Debug                      bool
}

type ReproductionMethod string

var (
	RouletteWheel               ReproductionMethod = "rouletteWheel"
	Tournament                  ReproductionMethod = "tournament"
	RankSelection               ReproductionMethod = "rankSelection"
	StochasticUniversalSampling ReproductionMethod = "stochasticUniversalSampling"
)

type Solver struct {
	Config            Config
	Population        []*genome.Genome
	BestGenome        *genome.Genome
	BestSolution      float64
	WorstSolution     float64
	AverageScore      float64
	Diversity         float64
	RNG               *rand.Rand
	OptLog            []OptimizationLog
	LatestImprovement int
	Generation        int
	Finished          bool
}

type OptimizationLog struct {
	Generation    int     `csv:"generation"`
	BestSolution  float64 `csv:"best_solution"`
	WorstSolution float64 `csv:"worst_solution"`
	AverageScore  float64 `csv:"average_score"`
	Diversity     float64 `csv:"diversity"`
}

func NewSolver(cfg Config) *Solver {
	return &Solver{
		Config: cfg,
		RNG:    rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}

func (s *Solver) Finish() {
	fmt.Printf("Finishing...\n")
	s.Finished = true
}

func (s *Solver) Optimize(debug bool) {
	if s.Population == nil || len(s.Population) == 0 {
		s.SeedPopulation()
	}
	// for generation := 0; generation-s.LatestImprovement < s.Config.GenerationImprovementLimit; generation++ {
	for generation := 0; ; generation++ {
		if s.Finished {
			break
		}
		s.Generation = generation
		s.EvaluatePopulation()
		s.RankPopulation(generation)
		s.CalculateDiversity()
		if debug {
			fmt.Printf("Generation %d:\t", generation)
			fmt.Printf("Best solution: %f\t", s.BestSolution)
			fmt.Printf("Worst solution: %f\t", s.WorstSolution)
			fmt.Printf("Average score: %f\t", s.AverageScore)
			fmt.Printf("Diversity: %f\n", s.Diversity)
		}
		s.OptLog = append(s.OptLog, OptimizationLog{
			Generation:    generation,
			BestSolution:  s.BestSolution,
			WorstSolution: s.WorstSolution,
			AverageScore:  s.AverageScore,
			Diversity:     s.Diversity,
		})
		if s.LatestImprovement-s.Generation > 1000 {
			// Stuck in local optima, need to randomize more
			s.Replace(s.Clone(s.Population), []*genome.Genome{}, []*genome.Genome{})
			s.Mutate()
			continue
		}
		cloningSelection := s.SelectForCloning()
		cloned := s.Clone(cloningSelection)
		crossoverSelection := s.SelectForCrossover()
		babies := s.Crossover(crossoverSelection)
		randomSize := int(math.Round(float64(s.Config.PopulationSize) / 10))
		newRandomGenomes := make([]*genome.Genome, randomSize)
		for i := 0; i < randomSize; i++ {
			newRandomGenomes[i] = genome.NewRandomGenome2(s.RNG, len(s.Config.Locations))
		}
		s.Replace(cloned, babies, newRandomGenomes)
		s.Mutate()
	}
}

func (s *Solver) SeedPopulation() {
	for i := 0; i < s.Config.PopulationSize; i++ {
		s.Population = append(s.Population, genome.NewRandomGenome(s.RNG, len(s.Config.Locations)))
	}
}

func (s *Solver) EvaluatePopulation() {
	// Evaluate each genome in the population concurrently using goroutines
	var wg sync.WaitGroup

	for _, gene := range s.Population {
		wg.Add(1) // Increment the WaitGroup counter
		go func(g *genome.Genome) {
			defer wg.Done() // Decrement the counter when the goroutine completes
			g.Evaluate(s.Config.Locations, *s.Config.MapData, *s.Config.GeneralGameData)
		}(gene)
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
		fmt.Printf("(Generation %d)\tNew best: %f\n", s.Generation, s.BestGenome.Score)
	}
	s.Population = append([]*genome.Genome{s.BestGenome.Copy()}, s.Population[:s.Config.PopulationSize-1]...)
	s.WorstSolution = s.Population[len(s.Population)-1].Score
	s.AverageScore = 0
	for _, genome := range s.Population {
		s.AverageScore += genome.Score
	}
	s.AverageScore /= float64(len(s.Population))
}

func (s *Solver) SelectForCrossover() []*genome.Genome {
	tenth := len(s.Population) / 10
	selection := make([]*genome.Genome, len(s.Population)-len(s.Population)/10-tenth)
	for i := 0; i < len(s.Population)-len(s.Population)/10-tenth; i++ {
		selection[i] = s.Population[i]
	}
	return selection
}

func (s *Solver) SelectForCloning() []*genome.Genome {
	selection := make([]*genome.Genome, len(s.Population)/10)
	for i := 0; i < len(s.Population)/10; i++ {
		selection[i] = s.Population[i]
	}
	return selection
}

func (s *Solver) SelectForReplacement() []*genome.Genome {
	popLen := len(s.Population) / 2
	toBeReplaced := make([]*genome.Genome, popLen)
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
		mutProb := s.Config.MutationProbability
		if s.Diversity < 0.4 {
			mutProb += 2 * (0.4 - s.Diversity)
		}
		if s.LatestImprovement-s.Generation > 1000 {
			genome.MutateNeighbors(1.0, s.Config.Locations) // Mutate extra hard if stuck in local optima
		} else if rand.Float64() < s.Config.MutationProbability {
			genome.MutateNeighbors(s.Config.MutationProbability, s.Config.Locations)
		}
	}
}

func (s *Solver) Crossover(population []*genome.Genome) []*genome.Genome {
	lenPop := len(population)
	randomOrder := s.RNG.Perm(lenPop)
	babies := make([]*genome.Genome, len(population))
	for i := 0; i < len(population); i += 2 {
		// babies[i], babies[i+1] = population[randomOrder[i]].Crossover(population[randomOrder[i+1]])
		babies[i], babies[i+1] = population[randomOrder[i]].CrossoverSinglePair(population[randomOrder[i+1]])
	}
	return babies
}

func (s *Solver) Clone(genomes []*genome.Genome) []*genome.Genome {
	clones := make([]*genome.Genome, len(genomes))
	for i := 0; i < len(genomes); i++ {
		clones[i] = genomes[i].Copy()
	}
	return clones
}

func (s *Solver) Replace(cloned, babies, randomGenes []*genome.Genome) {
	newPopulation := make([]*genome.Genome, len(s.Population))
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

func (s *Solver) GetSolution() scoring.ScoredSolution {
	genome := s.BestGenome
	genomeLocation := make(map[string]types.LocationSolution)
	for j, loc := range s.Config.Locations {
		genomeLocation[loc.Name] = types.LocationSolution{
			Location: *loc,
			F3:       genome.Pairs[j].F3,
			F9:       genome.Pairs[j].F9,
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
	scoredSolution, err := scoring.CalculateScore(filtered, s.Config.MapData.Name, *s.Config.GeneralGameData, s.Config.Locations)
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
	header := []string{"generation", "best_solution", "worst_solution", "average_score", "diversity"}
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
			strconv.FormatFloat(logEntry.Diversity, 'f', -1, 64),
		}

		if err := writer.Write(record); err != nil {
			log.Fatal(err)
		}
	}
	return nil
}

func (s *Solver) WriteSolutionToFile(gameID string) error {
	file, err := os.Create(fmt.Sprintf("solutions/solution-%s-%s-%.2f.csv", s.Config.MapData.Name, time.Now().Format("2006-01-02-15-04-05"), s.BestSolution))
	if err != nil {
		panic(err)
	}
	defer file.Close()
	/* Write the solution as JSON to a file */
	bestGenome := s.BestGenome
	type SolutionInput struct {
		LocationName string `json:"locationName"`
		F3           int    `json:"freestyle3100Count"`
		F9           int    `json:"freestyle9100Count"`
	}
	solutions := make([]SolutionInput, 0)
	for i, loc := range s.Config.Locations {
		sol := SolutionInput{
			LocationName: loc.Name,
			F3:           bestGenome.Pairs[i].F3,
			F9:           bestGenome.Pairs[i].F9,
		}
		solutions = append(solutions, sol)
	}
	solution := struct {
		GameID    string          `json:"gameId"`
		MapName   string          `json:"mapName"`
		Locations []SolutionInput `json:"locations"`
	}{
		GameID:    gameID,
		MapName:   s.Config.MapData.Name,
		Locations: solutions,
	}
	err = json.NewEncoder(file).Encode(solution)
	if err != nil {
		return err
	}
	return nil
}

func (s *Solver) LoadSolutionFromFile(file *os.File) error {
	/* Load a solution from a file */
	type SolutionInput struct {
		LocationName string `json:"locationName"`
		F3           int    `json:"freestyle3100Count"`
		F9           int    `json:"freestyle9100Count"`
	}
	var solution struct {
		GameID    string          `json:"gameId"`
		MapName   string          `json:"mapName"`
		Locations []SolutionInput `json:"locations"`
	}
	err := json.NewDecoder(file).Decode(&solution)
	if err != nil {
		return err
	}

	pairs := make([]genome.Pair, len(s.Config.Locations))
	for _, p := range solution.Locations {
		for i, loc := range s.Config.Locations {
			if loc.Name == p.LocationName {
				pairs[i] = genome.Pair{
					F3: p.F3,
					F9: p.F9,
				}
			}
		}
	}
	g := genome.Genome{
		Pairs: pairs,
	}
	s.BestGenome = &g
	population := make([]*genome.Genome, s.Config.PopulationSize)
	population[0] = &g
	for i := 1; i < s.Config.PopulationSize; i++ {
		population[i] = g.Copy()
	}
	s.Population = population
	return nil
}

type (
	StationCounts struct {
		Zero float64
		One  float64
		Two  float64
	}
	DiversityPair struct {
		F3 StationCounts
		F9 StationCounts
	}
)

var sqrt3 = math.Sqrt(3)

func (s *Solver) CalculateDiversity() {
	// calculate diversity
	geneCount := make([]DiversityPair, len(s.Config.Locations))

	for _, genome := range s.Population {
		for i, pair := range genome.Pairs {
			switch pair.F3 {
			case 0:
				geneCount[i].F3.Zero++
			case 1:
				geneCount[i].F3.One++
			case 2:
				geneCount[i].F3.Two++
			}
			switch pair.F9 {
			case 0:
				geneCount[i].F9.Zero++
			case 1:
				geneCount[i].F9.One++
			case 2:
				geneCount[i].F9.Two++
			}
		}
	}
	popSize := float64(s.Config.PopulationSize)
	for i := 0; i < len(geneCount); i++ {
		geneCount[i].F3.Zero /= popSize
		geneCount[i].F3.One /= popSize
		geneCount[i].F3.Two /= popSize
		geneCount[i].F9.Zero /= popSize
		geneCount[i].F9.One /= popSize
		geneCount[i].F9.Two /= popSize
	}

	diversityShannon := make([]float64, 0, len(geneCount)*2)
	for _, pair := range geneCount {
		diversityShannon = append(diversityShannon, ShannonDiversityIndex(pair.F3))
		diversityShannon = append(diversityShannon, ShannonDiversityIndex(pair.F9))
	}

	totalDiversity := 0.0
	for _, shannon := range diversityShannon {
		totalDiversity += shannon
	}
	diversity := totalDiversity / float64(len(diversityShannon))

	s.Diversity = diversity / sqrt3
}

func ShannonDiversityIndex(counts StationCounts) float64 {
	prop := []float64{counts.Zero, counts.One, counts.Two}
	H := 0.0
	for _, p := range prop {
		if p == 0 {
			continue
		}
		H += p * math.Log2(p)
	}
	return -H
}
