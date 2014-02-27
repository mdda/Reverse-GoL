// An implementation of Conway's Game of Life.
// See reverse-gol.go for build/run

package main

import (
	"math/rand"
	"fmt"
	"runtime"
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

type IndividualResult struct {
	individual *Individual
	mismatch_from_true_start_initial, mismatch_from_true_start_final int
	mismatch_from_true_end_initial, mismatch_from_true_end_final int
	iter int
}

func create_solution(problem LifeProblem, lps *LifeProblemSet) *IndividualResult {
	// Create a population of potential boards
	pop_size := 1000
	pop := NewPopulation(pop_size, problem.steps, problem.end, &lps.transition_collection[problem.steps])
	for i:=0; i<pop_size; i++ {
		// Create a candidate starting point
		// NB:  We can only work from the problem.end
		pop.individual[i].start.CopyFrom(problem.end)
	}
	
	p_temp := NewPopulation(pop_size, problem.steps, problem.end, &lps.transition_collection[problem.steps])

	l := NewBoardIterator(board_width, board_height)
	
	var best_individual *Individual
	best_individual_start := NewBoard_BoolPacked(board_width, board_height)
	
	mismatch_from_true_start_initial, mismatch_from_true_start_latest := 0,0
	mismatch_from_true_end_initial, mismatch_from_true_end_latest := 0,0
	
	iter_max  := 2000
	iter_last := 0
	for iter:=0; iter<iter_max; iter++ {
		// Evaluate fitness of every individual in pop
		for i, individual := range pop.individual {
			l.current.CopyFrom(individual.start)
			
			mismatch_from_true_start:=999
			if lps.is_training {
				diff     := NewBoard_BoolPacked(board_width, board_height)
				mismatch_from_true_start = l.current.CompareTo(problem.start, diff)
				
				if i==0 { // NB: Best individual is always in [0] (forced there in GenerationAfter)
					mismatch_from_true_start_latest  = mismatch_from_true_start
					if iter == 0 { 
						mismatch_from_true_start_initial  = mismatch_from_true_start
					}
				}
			}
			
			l.Iterate(problem.steps)
			
			// This is 'allowed' since we know the end result, and can store the diff
			mismatch_from_true_end := l.current.CompareTo(problem.end, individual.diff)
			
			if i==0 { // NB: Best individual is always in [0] (forced there in GenerationAfter)
				mismatch_from_true_end_latest  = mismatch_from_true_end
				if iter == 0 { 
					mismatch_from_true_end_initial  = mismatch_from_true_end
				}
			}
			
			// This is a lower factor pressure, but good to have too
			count_on := individual.start.CompareTo(board_empty, nil)
			
			individual.fitness = -mismatch_from_true_end  -count_on*0
			//individual.fitness = -mismatch_from_true_end*4 -count_on*1
			//individual.fitness = -mismatch_from_true_end*problem.steps -count_on*1
			
			if i<3 && (iter % 100 == 0) {
				fmt.Printf("%4d.%3d : Mismatch vs true {start,end} = {%3d,%3d}\n", iter, i, mismatch_from_true_start, mismatch_from_true_end) // , individual.start
			}
			
			iter_last=iter
		}
		
		best_individual = pop.BestIndividual()
		//fmt.Printf("%4d.best: Mismatch vs true {start,end} = {???,%3d}\n", iter, best_individual.fitness)
		//fmt.Print(best_individual.start)

		if iter>0 && (iter % 100 == 0) {
			difference_over_100_generations := best_individual.start.CompareTo(best_individual_start, nil)
			if difference_over_100_generations == 0 {
				// Our best candidate hasn't improved in 100 generations
				// So our job is done!
				iter = iter_max
				break
			}
		}
		
		if iter % 100 == 0 {
			best_individual_start.CopyFrom(best_individual.start)
		}
		
		p_temp.GenerationAfter(pop)
		pop, p_temp = p_temp, pop // Switcheroo to advance to next population
	}
	
	return &IndividualResult{
		individual : best_individual, 
		
		mismatch_from_true_start_initial : mismatch_from_true_start_initial, 
		mismatch_from_true_start_final   : mismatch_from_true_start_latest,
		mismatch_from_true_end_initial : mismatch_from_true_end_initial, 
		mismatch_from_true_end_final   : mismatch_from_true_end_latest,
		
		iter:iter_last,
	}
}

// http://devcry.heiho.net/2012/07/golang-masterworker-in-go.html
type Work struct {
	id int
	i,n int
	is_training bool
	steps int
	lps *LifeProblemSet
}

func problem_worker_for_queue(worker_id int, queue chan *Work) {
	var wp *Work
	for {
		// get work item (pointer) from the queue
		wp = <-queue
		if wp == nil {
			break
		}
		fmt.Printf("worker #%d: item %v\n", worker_id, *wp)
		
		id := wp.id

		seed := get_unprocessed_seed_from_db(id, wp.is_training)
		fmt.Printf("(%5d/%5d) Running problem[%d].steps=%d (seed=%d)\n", wp.i, wp.n, id, wp.steps, seed)
		rand.Seed(int64(seed))
		individual_result := create_solution(wp.lps.problem[id], wp.lps)
		save_solution_to_db(id, wp.steps, seed, individual_result, wp.is_training)
	}
}

func pick_problems_from_db_and_solve_them(steps int, problem_count_requested int, is_training bool) {  
	var kaggle LifeProblemSet
	
	problem_list := list_of_interesting_problems_from_db(steps, problem_count_requested, is_training)
	
	//problem_list := []int{50,54}
	kaggle.load_csv(is_training, problem_list)

	// Now ensure that the transition_collection is valid for this step size
	kaggle.load_transition_collection(steps)
	
	queue := make(chan *Work)

	n_problems := len(problem_list)

	ncpu := runtime.NumCPU()
	if n_problems < ncpu {
		ncpu = n_problems
	}
	runtime.GOMAXPROCS(ncpu)

	// spawn workers
	for i := 0; i < ncpu; i++ {
		go problem_worker_for_queue(i, queue)
	}

	// master: give work
	for i, id := range problem_list {
		if kaggle.problem[id].steps != steps {
			fmt.Printf("Need to match problem[%d].steps=%d (not %d)\n", id, kaggle.problem[id].steps, steps)
		}
		wp := Work{
			id:id, 
			i:i, n:n_problems,
			is_training:is_training,
			steps:steps, 
			lps:&kaggle,
		}
		
		fmt.Printf("master: give work %v\n", wp)
		//queue <- &work[i]  // be sure not to pass &item !!!
		queue <- &wp  // be sure not to pass &item !!!
	}

	// all work is done
	// push ncpu*nil on the queue so that each worker will receive signal that there is no more work
	for n := 0; n < ncpu; n++ {
		queue <- nil
	}
}

