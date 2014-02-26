// An implementation of Conway's Game of Life.
// See reverse-gol.go for build/run

package main

import (
	"math/rand"
//	"fmt"
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
	target *Board_BoolPacked
	
	pressure_pct int
	
	mutation_pct int // (0..100)
	mutation_radius int
	mutation_loop_pct int

	crossover_pct int // (0..100)
	
	transition_collection *TransitionCollectionList
}

func NewPopulation(size int, radius int, target *Board_BoolPacked, tc *TransitionCollectionList) *Population {
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
		target:target,
		transition_collection:tc,
		
		pressure_pct:1*90+0*100,  // pressure_pct is in (50..100) = Prob(Chose better of two random individuals)
		
		mutation_pct:1*50+0*100,
		mutation_loop_pct:70,
		mutation_radius:radius,
		
		crossover_pct:30*1,
	}
}

func (p *Population) OrderIndividualsBasedOnFitness(i_1, i_2 *Individual) (*Individual,*Individual) {  
	if i_1.fitness == i_2.fitness {
		// Secondary pressure to minimize count of on cells in starting board
		
		count_on_1 := i_1.start.CompareTo(board_empty, nil)
		count_on_2 := i_2.start.CompareTo(board_empty, nil)
		
		if count_on_1 > count_on_2 {
			i_1,i_2 = i_2,i_1 // Switch them so that i_1 is the fitter (lower is better for initial positions) of the two
		}
	} else if i_1.fitness < i_2.fitness {
		i_1,i_2 = i_2,i_1 // Switch them so that i_1 is the fitter (higher is better) of the two
	}
	return i_1, i_2
}

func (p *Population) PickIndividualWithPressure() *Individual {  
	// Pick two individuals at random from population
	i_1_pos := rand.Intn(len(p.individual))
	i_1 := p.individual[i_1_pos]
	
	i_2_pos := rand.Intn(len(p.individual))
	i_2 := p.individual[i_2_pos]
	
	i_1, i_2 = p.OrderIndividualsBasedOnFitness(i_1, i_2)
	
	// if pct< a threshold, pick the better one
	i_chosen := i_1
	if rand.Intn(100) > p.pressure_pct { // i.e. only sometimes do the opposite
		i_chosen = i_2
	}
	//fmt.Printf("Individuals {%d:%d} Fitnesses : {%d:%d} -> %d\n", i_1_pos, i_2_pos, i_1.fitness, i_2.fitness, i_chosen.fitness)
	
	return i_chosen
}

func (pop *Population) BestIndividual() *Individual {
	i_best := pop.individual[0]
	for _,i := range pop.individual {
		i_best,_ = pop.OrderIndividualsBasedOnFitness(i_best, i)
	}
	return i_best
}

func (pop *Population) GenerationAfter(prev *Population) {
	// Fill in every slot
	for counter, individual := range pop.individual {
		if counter==0 { // Reserve position 0 for a copy of the previous generation's best individual
			best_individual := prev.BestIndividual()
			individual.start.CopyFrom(best_individual.start)
			individual.fitness = 0
			continue
		}
		
		choser := rand.Intn(100)
		if 0<=choser && choser < pop.crossover_pct { 
			// Do a 'crossover copy' from two individuals in previous population to this one
			parent_1 := prev.PickIndividualWithPressure()
			parent_2 := prev.PickIndividualWithPressure()
			individual.start.CrossoverFrom(parent_1.start, parent_2.start)
		} else { // Do a simple copy, with the possibility of mutation (below)
			i_chosen := prev.PickIndividualWithPressure()
			individual.start.CopyFrom(i_chosen.start)
			if pop.crossover_pct<=choser && choser < (pop.crossover_pct + pop.mutation_pct) {
				//individual.start.MutateRadiusBits(pop.mutation_loop_pct, pop.mutation_radius) // % do additional mutation, radius of action
				
				// For this individual, pick a position in the diff
				x,y := i_chosen.diff.RandomBitPosition()
				
				if x>=0 && y>=0 {
					// Offset by a little bit...
					if true {
						//fmt.Printf("target_error@(%2d,%2d):\n", x,y)
						x = CoordWithinRadius(x, i_chosen.diff.w, pop.mutation_radius/2+1)
						y = CoordWithinRadius(y, i_chosen.diff.h, pop.mutation_radius/2+1)
					}
				} else {
					// There are no errors...  So we don't have a basis for complaining, really
					// So let's create a mask based on bits from the end (for variety)
					
					if false {
						//fmt.Printf("No errors to mutate around : Try using the target instead of the diff\n")
						//fmt.Println(i_chosen.start) // Check
						x,y = pop.target.RandomBitPosition()
						//fmt.Println("*** Isn't the end image DEFINED to be non-blank? ***")
					}
					
					if true {
						//fmt.Printf("No errors to mutate around : Try zeroing out bits in the start\n")
						individual.start.MutateMask(individual.start, pop.mutation_loop_pct, 0) // % do additional mutation, radius of action
						x,y = -1,-1 // Don't do the overlay thing
					}
				}
				if x>=0 && y>=0 {
					end := pop.target.MakePatch(x,y)
					//fmt.Printf("Examining patch(%8d) from target @(%2d,%2d):\n", int(end), x,y)
					//fmt.Print(end)
					
					start_random := pop.transition_collection.GetRandomEntry_OrientationCompensated(end)
					if start_random>=0 { // Yes - we have an overlay to try...
						//fmt.Print("Suggested Start :\n")
						//fmt.Print(start_random)
						individual.start.OverlayPatch(x,y, start_random)
					} else {
						//fmt.Printf("Did not find known start patch for :\n")
						//fmt.Print(end)
						
						// Use the diff mask calculated for the chosen individual instead (i.e. DO SOMETHING)
						// individual.start.MutateMask(i_chosen.diff, pop.mutation_loop_pct, pop.mutation_radius)
						
						// Use the chosen individual bits as a bit mask instead (i.e. DO SOMETHING)
						
						if true {
							//fmt.Printf("Introducing random noise\n")
							individual.start.MutateMask(individual.start, pop.mutation_loop_pct, pop.mutation_radius)
						}
						
						if false {
							//fmt.Printf("Introducing random zeroing\n")
							individual.start.MutateMask(individual.start, pop.mutation_loop_pct, 0) // % do additional mutation, radius of action
						}
					}
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
