// An implementation of Conway's Game of Life.
// See reverse-gol.go for build/run

package main

import (
	"math/rand"
)

type Individual struct {
	start *Board_BoolPacked
	diff  *Board_BoolPacked
	fitness int  // higher is better, no particular scale
}

/*
func IndividualMutationGenerator(amount int) func f(i *Individual) *Individual{
	return func(i *Individual) *Individual{
		
	}
}
*/

type Population struct {
	individual []*Individual
	
	pressure_pct int
	
	mutation_pct int // (0..100)
	mutation_radius int
	mutation_loop_pct int

	crossover_pct int // (0..100)
	
	lps *LifeProblemSet // This is reference to the basic LifeProblemSet we're working on (including the transition stats)
}

func NewPopulation(size int, radius int, lps *LifeProblemSet) *Population {
	//fmt.Printf("NewPopulation(size=%d)\n", size)
	ind := make([]*Individual, size)
	for i:=0; i<size; i++ {
		ind[i] = &Individual{ 
			                  start:NewBoard_BoolPacked(board_width, board_height), 
			                  diff: NewBoard_BoolPacked(board_width, board_height), 
		                      fitness:0,
		                    }
	}
	//fmt.Printf("NewPopulation(size=%d) inited\n", size)
	return &Population{
		individual:ind,
		
		pressure_pct:1*70+0*100,  // pressure_pct is in (50..100) = Prob(Chose better of two random individuals)
		
		mutation_pct:1*30+0*100,
		mutation_loop_pct:70,
		mutation_radius:radius,
		
		crossover_pct:30*1,
		
		lps:lps,
	}
}

func (p *Population) PickIndividualWithPressure() *Individual {  
	// Pick two individuals at random from population
	i_1_pos := rand.Intn(len(p.individual))
	i_1 := p.individual[i_1_pos]
	
	i_2_pos := rand.Intn(len(p.individual))
	i_2 := p.individual[i_2_pos]
	
	if i_1.fitness < i_2.fitness {
		i_1,i_2 = i_2,i_1 // Switch them so that i_1 is the fitter (higher is better) of the two
	}
	
	// if pct< a threshold, pick the better one
	i_chosen := i_1
	if rand.Intn(100) > p.pressure_pct { // i.e. only sometimes do the opposite
		i_chosen = i_2
	}
	//fmt.Printf("Individuals {%d:%d} Fitnesses : {%d:%d} -> %d\n", i_1_pos, i_2_pos, i_1.fitness, i_2.fitness, i_chosen.fitness)
	
	return i_chosen
}

func (p *Population) GenerationAfter(prev *Population) {
	// Fill in every slot
	for _, individual := range p.individual {
		choser := rand.Intn(100)
		if 0<=choser && choser < p.crossover_pct { 
			// Do a 'crossover copy' from two individuals in previous population to this one
			parent_1 := prev.PickIndividualWithPressure()
			parent_2 := prev.PickIndividualWithPressure()
			individual.start.CrossoverFrom(parent_1.start, parent_2.start)
		} else { // Do a simple copy, with the possibility of mutation (below)
			i_chosen := prev.PickIndividualWithPressure()
			individual.start.CopyFrom(i_chosen.start)
			if p.crossover_pct<=choser && choser < (p.crossover_pct + p.mutation_pct) {
				//individual.start.MutateRadiusBits(p.mutation_loop_pct, p.mutation_radius) // % do additional mutation, radius of action
				
				// Use the diff mask calculated for the chosen individual
				//individual.start.MutateMask(i_chosen.diff, p.mutation_loop_pct, p.mutation_radius)
				
				// For this individual, pick a position in the diff
				x,y := i_chosen.diff.RandomBitPosition()
				if x>=0 && y>=0 {
					// Now, find that thing in the TransitionMap 
					
					// if found, then copy a random one of its starters into the new individual
					
					
					//start_random
					//individual.start.OverlayPatch(x,y, start_random)
				}
				
			}
		}

		individual.fitness = 0
	}
}

/*
func (p *Population) MutateIndividuals(another_mutation_pct int, radius int) {  // % do additional mutation, radius of action
	for c, individual := range p.individual {
		//individual.start.MutateFlipBits(rand.Intn(mutation_size))
		individual.start.MutateRadiusBits(another_mutation_pct, radius)
	}
}
*/
