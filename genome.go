package main

import (
	"math"
	"math/rand"
)

type Pair struct {
	F3 int
	F9 int
}

type Genome struct {
	Pairs []Pair
	Score float64
}

func NewRandomGenome(rng *rand.Rand, locations int) *Genome {
	pairs := make([]Pair, locations)
	for i := 0; i < locations; i++ {
		f3 := rng.Intn(6)
		f9 := rng.Intn(6)
		pairs[i] = Pair{
			F3: f3,
			F9: f9,
		}
	}
	return &Genome{
		Pairs: pairs,
	}
}

func (g *Genome) Fitness() float64 {
	return g.Score
}

func (g *Genome) Copy() *Genome {
	c := &Genome{
		Pairs: make([]Pair, len(g.Pairs)),
		Score: g.Score,
	}
	copy(c.Pairs, g.Pairs)
	return c
}

func (g *Genome) Mutate(mutationProb float64) {
	for i := range g.Pairs {
		if rand.Float64() < mutationProb {
			negative := rand.Float64() < 0.5
			if negative {
				newVal := math.Max(0, float64(g.Pairs[i].F3-1))
				g.Pairs[i].F3 = int(newVal)
			} else {
				newVal := math.Min(5, float64(g.Pairs[i].F3+1))
				g.Pairs[i].F3 = int(newVal)
			}
		}
		if rand.Float64() < mutationProb {
			negative := rand.Float64() < 0.5
			if negative {
				newVal := math.Max(0, float64(g.Pairs[i].F9-1))
				g.Pairs[i].F9 = int(newVal)
			} else {
				newVal := math.Min(5, float64(g.Pairs[i].F9+1))
				g.Pairs[i].F9 = int(newVal)
			}
		}
	}
}

func (g *Genome) Crossover(other *Genome) (*Genome, *Genome) {
	crossoverPoint := rand.Intn(len(g.Pairs))
	c1Pairs := append(g.Pairs[:crossoverPoint], other.Pairs[crossoverPoint:]...)
	c2Pairs := append(other.Pairs[:crossoverPoint], g.Pairs[crossoverPoint:]...)
	c1 := &Genome{
		Pairs: c1Pairs,
	}
	c2 := &Genome{
		Pairs: c2Pairs,
	}
	return c1, c2
}

func (g *Genome) Evaluate(locations []*Location, mapData MapData, generalData GeneralGameData) {
	genomeLocation := make(map[string]LocationSolution)
	for j, loc := range locations {
		genomeLocation[loc.Name] = LocationSolution{
			Location: *loc,
			F3:       g.Pairs[j].F3,
			F9:       g.Pairs[j].F9,
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
		g.Score = 0
		return
	}
	scoredSolution, err := CalculateScore(filtered, mapData, generalData)
	if err != nil {
		panic(err)
	}
	g.Score = scoredSolution.GameScore["total"]

}
