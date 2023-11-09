package main

import "math/rand"

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
			g.Pairs[i].F3 = rand.Intn(6)
		}
		if rand.Float64() < mutationProb {
			g.Pairs[i].F9 = rand.Intn(6)
		}
	}
}
