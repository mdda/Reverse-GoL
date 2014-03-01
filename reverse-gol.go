package main

// GOPATH=`pwd` go build reverse-gol.go speed_packed.go ga.go board-standard.go transitions.go db.go && ./reverse-gol

import (
	"fmt"
	"time"
	"math/rand"
	"flag"
)


func main_orig() {
	l := NewBoardIterator(board_width, board_height)
	for i := 0; i < 65; i++ {
		l.Iterate(1)
		fmt.Print("\x0c", l) // Clear screen and print field.
		time.Sleep(time.Second / 30)
	}
}

func main_test_random() {
	for pct := float32(0.1); pct < 1.0; pct += 0.1 {
		l := NewBoardIterator(board_width, board_height)
		l.current.UniformRandom(pct + 0.01)

		//fmt.Print("\x0c", l)
		fmt.Print(l.current, "\n")
		l.Iterate(5) // 65 max reasonable steps for glider
		fmt.Print(l.current, "\n")
	}
}

func main_timer() {
	const glider = `
--X
X-X
-XX`

	start := time.Now()

	for iter := 0; iter < 1000; iter++ {
		l := NewBoardIterator(board_width, board_height)
		l.current.LoadString(glider[1:])

		l.Iterate(65)

		if iter == 0 {
			fmt.Print(l.current, "\n")
		}
	}

	elapsed := time.Since(start)
	fmt.Printf("1000 iterations took %s\n", elapsed)
}

func main_loader() {
	for pct := float32(0.1); pct < 1.0; pct += 0.1 {
		l := NewBoardIterator(board_width, board_height)
		l.current.UniformRandom(pct + 0.01)

		//fmt.Print("\x0c", l)
		fmt.Print(l.current, "\n")
		l.Iterate(5)
		fmt.Print(l.current, "\n")
	}
}

func main_verify_training_examples(problem_offset int) {
	var kaggle LifeProblemSet
	is_training := true // only makes sense on training_data...

	id_list := []int{}
	for id := problem_offset; id < problem_offset+10; id++ {
		id_list = append(id_list, id)
	}
	kaggle.load_csv(is_training, id_list)
	//fmt.Println(kaggle.problem[107].start)
	//fmt.Println(kaggle.problem[107].end)

	image := NewImageSet(10, 11) // 10 rows of 11 images each, formatted 'appropriately'

	for _, id := range id_list {
		bs_start := NewBoardStats(board_width, board_height)
		kaggle.problem[id].start.AddToStats(bs_start)

		bs_end := NewBoardStats(board_width, board_height)
		kaggle.problem[id].end.AddToStats(bs_end)

		//image.DrawStats(r,c, bs)
		image.DrawStatsNext(bs_start)
		image.DrawStats(image.row_current, image.cols-1, bs_end)

		l := NewBoardIterator(board_width, board_height)
		l.current.CopyFrom(kaggle.problem[id].start)
		
		//fmt.Printf("Training Example[%d].steps=%d\n", id, steps)

		for i := 0; i < kaggle.problem[id].steps; i++ {
			l.Iterate(1) // Just 1 step per image for now

			bs_now := NewBoardStats(board_width, board_height)
			l.current.AddToStats(bs_now)
			image.DrawStatsNext(bs_now)
		}
		image.DrawStatsNext(bs_end) // For ease of comparison..
	
		if mismatch := l.current.CompareTo(kaggle.problem[id].end, nil); mismatch>0 {
			fmt.Printf("** Training Example[%d] FAIL - by %d\n", id, mismatch)
		}

		image.DrawStatsCRLF()
	}

	image.save("images/train.png")
}

func main_visualize_density() {
	image := NewImageSet(10, 11) // 10 rows of 11 images each, formatted 'appropriately'

	for pct:=float32(0.1); pct<0.99; pct+=0.1 {
		var bs []*BoardStats
		for j:=0; j<10; j++ {
			bs = append(bs, NewBoardStats(board_width, board_height))
		}

		for i:=0; i<1000; i++ {
			l := NewBoardIterator(board_width, board_height)
			l.current.UniformRandom(pct)
			l.current.AddToStats(bs[0])
			for j:=1; j<len(bs); j++ {
				l.Iterate(1)
				l.current.AddToStats(bs[j])
			}
		}
		for j:=0; j<len(bs); j++ {
			image.DrawStatsNext(bs[j])
		}
		image.DrawStatsCRLF()
	}

	image.save("images/density.png")
}

func main_population_score(is_training bool, id int) {
	image := NewImageSet(10, 12) // 10 rows of 12 images each, formatted 'appropriately'
	
	var kaggle LifeProblemSet
	kaggle.load_csv(is_training, []int{id}) // Load from the training set

	problem := kaggle.problem[id]
	
	// Now ensure that the transition_collection is valid for this step size
	kaggle.load_transition_collection(problem.steps)
	
	// This is the TRUE starting place : for reference
	bs_start := NewBoardStats(board_width, board_height)
	problem.start.AddToStats(bs_start)

	// This is the TRUE ending place : for reference
	bs_end := NewBoardStats(board_width, board_height)
	problem.end.AddToStats(bs_end)

	// Create a population of potential boards
	pop_size := 1000
	pop := NewPopulation(pop_size, problem.steps, problem.end, &kaggle.transition_collection[problem.steps])
	for i:=0; i<pop_size; i++ {
		// Create a candidate starting point
		// NB:  We can only work from the problem.end
		pop.individual[i].start.CopyFrom(problem.end)
		//pop.individual[i].start.UniformRandom(0.32)
	}
	
	p_temp := NewPopulation(pop_size, problem.steps, problem.end, &kaggle.transition_collection[problem.steps])

	l := NewBoardIterator(board_width, board_height)
	
	iter_max := 1000
	for iter:=0; iter<iter_max; iter++ {
		disp_row := (0 == (iter) % (iter_max/10))
		
		if disp_row {
			// for ease of comparison
			image.DrawStatsNext(bs_start)
		}
		
		// Evaluate fitness of every individual in pop
		//for i:=0; i<pop_size; i++ {
		for i, individual := range pop.individual {
			l.current.CopyFrom(individual.start)
			
			//l.Iterate(5)
			
			mismatch_from_true_start:=999
			if i<5 && disp_row {
				diff     := NewBoard_BoolPacked(board_width, board_height)
				mismatch_from_true_start = l.current.CompareTo(problem.start, diff)
				
				bs_trial := NewBoardStats(board_width, board_height)
				l.current.AddToStats(bs_trial)
				//diff.AddToStats(bs_trial)
				
				//fmt.Printf("\n%3d.%2d : Mismatch from true start = %d\n", iter, i, mismatch_from_true_start)
				bs_trial.MisMatchBy(mismatch_from_true_start)
			
				image.DrawStatsNext(bs_trial)
			}
			
			l.Iterate(problem.steps)
			
			// This is 'allowed' since we know the end result, and can store the diff
			mismatch_from_true_end := l.current.CompareTo(problem.end, individual.diff)
			
			// This is a lower factor pressure, but good to have too
			count_on := individual.start.CompareTo(board_empty, nil)
			
			//individual.fitness = -mismatch_from_true_end
			//individual.fitness = -mismatch_from_true_end*4 -count_on*1
			individual.fitness = -mismatch_from_true_end*problem.steps -count_on*1
			
			if i<5 && disp_row {
				bs_result := NewBoardStats(board_width, board_height)
				//l.current.AddToStats(bs_result)
				individual.diff.AddToStats(bs_result)
				
				show_best:=""
				if i==0 {
					show_best=" <<< best"
				}
				fmt.Printf("%4d.%3d : Mismatch vs true {start,end} = {%3d,%3d}%s\n", iter, i, mismatch_from_true_start, mismatch_from_true_end, show_best) // , individual.start
				bs_result.MisMatchBy(mismatch_from_true_end)
				
				image.DrawStatsNext(bs_result)
			}
			
			if mismatch_from_true_end==0 {
				//fmt.Printf("%4d.%3d : Mismatch vs true {start,end} = {%3d,%3d} :: PERFECTION!\n", iter, i, mismatch_from_true_start, mismatch_from_true_end)
			}
		}
		
		if disp_row {
			//image.DrawStatsNext(bs_end) // For ease of comparison..
			image.DrawStats(image.row_current, image.cols-1, bs_end)
			
			image.DrawStatsCRLF()
		}
		
		/*
		best_individual := pop.BestIndividual()
		fmt.Printf("%4d.best: Mismatch vs true {start,end} = {???,%3d}\n", iter, best_individual.fitness)
		fmt.Print(best_individual.start)
		
		if best_individual.fitness>=0 {
			// We have solved it, really
			//fmt.Print(best_individual.start)
			//break
		}
		*/
		
		p_temp.GenerationAfter(pop)
		pop, p_temp = p_temp, pop // Switcheroo to advance to next population
	}

	//image.DrawStats(image.row_current, image.cols-1, bs_end)
	
	image.save("images/score_mutated.png")
}

func main_create_stats(steps int) {
	var transitions TransitionCollectionMap
	
	transitions.TrainingCSV_to_stats("data/train.csv", steps) 
	// No flipping stats (#steps, count_of_training_examples, unique end-point-patches_raw, unique end-point-patches_ud-or-lr, unique end-point-patches_ud*lr ):
	// 1  9866 648k 471k 454k
	// 2 10042 620k 450k 434k
	// 3  9947 589k 427k 411k
	// 4 10089 565k 410k 394k
	// 5  9956 534k 387k 374k
	
	transitions.SaveCSV(fmt.Sprintf(TransitionCollectionFileStrFmt, steps))
}

func main_create_stats_all() {
	for _,i := range( []int{1,2,3,4,5} ) {
		main_create_stats(i)
	}
	/*
[andrewsm@square reverse-gol]$ ls -l stats/
total 66520
-rw-rw-r--. 1 andrewsm andrewsm 12129025 Feb 27 01:09 transition-1.csv
-rw-rw-r--. 1 andrewsm andrewsm 13549146 Feb 27 01:09 transition-2.csv
-rw-rw-r--. 1 andrewsm andrewsm 14002885 Feb 27 01:09 transition-3.csv
-rw-rw-r--. 1 andrewsm andrewsm 14329277 Feb 27 01:10 transition-4.csv
-rw-rw-r--. 1 andrewsm andrewsm 14098291 Feb 27 01:10 transition-5.csv
	 */
}


func main_read_stats(steps int) {
	var transitions TransitionCollectionList
	
	transitions.LoadCSV(fmt.Sprintf(TransitionCollectionFileStrFmt, steps)) 
}

func main_create_fake_training_data() {
	lps := LifeProblemSet{
		problem:make(map[int]LifeProblem),
		is_training:true,
	}
	
	num_per_step:=200
	
	base := 60*1000 + 1 // i.e. 60k+1 onwards
	for steps:=1; steps<=5; steps++ {
		for id:=base; id<base+num_per_step; id++ {
			problem := LifeProblem{id:id, steps:steps}
			problem.CreateFake()
			lps.problem[id]=problem
		}
		base+=num_per_step
	}
	lps.save_csv("data/train_fake.csv")
}

const currently_running_version int = 1002

func main() {
	cmd:= flag.String("cmd", "", "Required : {db|create|visualize|run|submit}")
	cmd_type:= flag.String("type", "", "create:{fake_training_data|training_set_transitions|synthetic_transitions}, db:{test|insert_problems}, visualize:{data|ga}")
	
	delta := flag.Int("delta", 0, "Number of steps between start and end")
	seed  := flag.Int64("seed", 1, "Random seed to use")

	id := flag.Int("id", 0, "Specific id to examine")
	training_only := flag.Bool("training", false, "Act on training set (default=false, i.e. test set)")

	count := flag.Int("count", 0, "Number of ids to process")

	
	flag.Parse()
	//fmt.Printf("CMD = %s\n", *cmd)
	
	//rand.Seed(time.Now().UnixNano()) 
	rand.Seed(*seed)
	
	//main_timer()
	//main_visualize_density()
	
	//main_verify_training_examples()
	
	//main_population_score()
	
	if *cmd=="db" {
		/// ./reverse-gol -cmd=db -type=test
		if *cmd_type=="test" {
			test_open_db()
		}
		
		/// ./reverse-gol -cmd=db -type=insert_problems
		if *cmd_type=="insert_problems" {
			create_list_of_problems_in_db() // NB: This sets up the 'problems' table to want answers...
		}
		
		//reset_all_currently_processing(-1)
		
		//probs := list_of_interesting_problems_from_db(1,5,true) // training 
		//fmt.Println(probs)
	}
	
	if *cmd=="create" {
		/// ./reverse-gol -cmd=create -type=fake_training_data
		if *cmd_type=="fake_training_data" {
			main_create_fake_training_data()
			
			// Prevent solving of actual training set (since this is where our state came from, so it's not particularly helpful
			// UPDATE problems SET solution_count=100 WHERE id>-60000  and id<0
			// UPDATE problems SET solution_count=0 WHERE id>-100000 and id<-60000
		}
		
		if *cmd_type=="training_set_transitions" {
			//main_create_stats(1)
			//main_create_stats_all()
			//main_read_stats(1)
		}
		
		if *cmd_type=="synthetic_transitions" {
			
		}
		
	}

	if *cmd=="visualize" {
		/// ./reverse-gol -cmd=visualize -type=data -training=true -id=50
		/// ./reverse-gol -cmd=visualize -type=data -training=true -id=60001
		/// ./reverse-gol -cmd=visualize -type=data -training=true -id=60201
		/// ./reverse-gol -cmd=visualize -type=data -training=true -id=60401
		/// ./reverse-gol -cmd=visualize -type=data -training=true -id=60601
		/// ./reverse-gol -cmd=visualize -type=data -training=true -id=60801
		if *cmd_type=="data" {
			if *id<=0 {
				fmt.Println("Need to specify '-id=%d' as base id to view (will also show 9 following)")
				flag.Usage()
				return
			}
			if !*training_only {
				fmt.Println("Need to specify '-training=true' (don't know start boards for test...)")
				flag.Usage()
				return
			}
			main_verify_training_examples(*id)
		}
		
		/// ./reverse-gol -cmd=visualize -type=ga -id=58 -training=true
		if *cmd_type=="ga" {
			if *id<=0 {
				fmt.Println("Need to specify '-id=%d'")
				flag.Usage()
				return
			}
			main_population_score(*training_only, *id)
		}
	}

	if *cmd=="run" {
		// To force DB to solve the 2 training problems : 50 and 54 : 
		// UPDATE problems SET solution_count=-2 WHERE id=-50 OR id=-54
		// ./reverse-gol -cmd=run -delta=5 -count=200 -training=true
		
		/// ./reverse-gol -cmd=run -delta=4 -count=10000
		if *delta<=0 {
			fmt.Println("Need to specify '-delta=%d'")
			flag.Usage()
			return
		}
		if *count<=0 {
			fmt.Println("Need to specify '-count=%d'")
			flag.Usage()
			return
		}
		if *training_only {
			fmt.Println("Running on Training Data")
		}
		problem_count_requested:=*count // This may be truncated, if there are less available ids (some may be processing already)
		steps := *delta
		pick_problems_from_db_and_solve_them(steps, problem_count_requested, *training_only)
	}
	
	if *cmd=="submit" {
		// create submission
		fname := fmt.Sprintf("submissions/submission_mdda_%s.csv", time.Now().Format("2006-01-02_15-04"))
		//fmt.Println(fname)
		create_submission(fname)
	}
	
	//fmt.Printf("Random #%3d\n", rand.Intn(1000))
}

