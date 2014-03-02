package main

// Installation of required library : 
// GOPATH=`pwd` go get github.com/go-sql-driver/mysql

// Create a database user 'reverse-gol' with password 'reverse-gol' with access rights to database 'reverse-gol'

import (
	"database/sql"
	_ "github.com/go-sql-driver/mysql"
	
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
)

/*
CREATE TABLE `problems` ( 
	`id` int(11) NOT NULL, 
	`steps` int(11) NOT NULL, 
	`solution_count` int(11) NOT NULL,
	`currently_processing` int(11) NOT NULL,
	KEY `problems_id` (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=latin1
CREATE TABLE `solutions` ( 
	`id` int(11) NOT NULL, 
	`steps` int(11) NOT NULL, 
	`iter` int(11) NOT NULL, 
	`seed` int(11) NOT NULL, 
	`version` int(11) NOT NULL, 
	`ones_i` int(11) DEFAULT NULL, 
	`mtsi` int(11) DEFAULT NULL, 
	`mtei` int(11) NOT NULL, 
	`ones_f` int(11) DEFAULT NULL, 
	`mtsf` int(11) DEFAULT NULL, 
	`mtef` int(11) NOT NULL, 
	`start` text NOT NULL, 
	KEY `solutions_id` (`id`) 
) ENGINE=InnoDB DEFAULT CHARSET=latin1
*/

func get_db_connection() *sql.DB {
	//db, err := sql.Open("mysql", "user:password@/database")
	//db, err := sql.Open("mysql", "reverse-gol:reverse-gol@/reverse-gol")
	db, err := sql.Open("mysql", "reverse-gol:reverse-gol@tcp(square.herald:3306)/reverse-gol")
	
	if err != nil {
		panic(err.Error()) // Just for example purpose. You should use proper error handling instead of panic
	}
	
	// Open doesn't open a connection. Validate DSN data:
	err = db.Ping()
	if err != nil {
		panic(err.Error()) // proper error handling instead of panic in your app
	}
	
	return db
}

func test_open_db() {
	db := get_db_connection()
	defer db.Close()
}

func create_list_of_problems_in_db() {
	db := get_db_connection()
	defer db.Close()
	
	//db.Exec("DELETE FROM PROBLEMS")
	
	sel, err := db.Prepare("SELECT id FROM problems WHERE id=?")
	if err != nil {
		fmt.Println("Select Prepare Error:", err)
		return
	}
	defer sel.Close()
	
	ins, err_ins := db.Prepare("INSERT INTO problems SET id=?, steps=?, solution_count=?")
	if err_ins != nil {
		fmt.Println("Insert Prepare Error:", err_ins)
		return
	}
	defer ins.Close()
	
	for _,f := range []string{"data/test.csv", "data/train.csv", "data/train_fake.csv", } {
		fmt.Printf("Opening %s\n", f)
		
		file, err := os.Open(f)
		if err != nil {
			fmt.Println("File Opening Error:", err)
			continue // If file not found, try the next one
		}
		defer file.Close()
		
		reader := csv.NewReader(file)
		
		header, err := reader.Read()
		if header[0] != "id" {
			fmt.Println("Bad Header", err)
			return
		}

		for {
			record, err := reader.Read()
			if err == io.EOF {
				break
			} else if err != nil {
				fmt.Println("Error:", err)
				return
			}

			id, _ := strconv.Atoi(record[0])
			steps, _ := strconv.Atoi(record[1])
			
			if !strings.HasSuffix(f, "/test.csv") {
				// This is training data : Let's give it a negative id to avoid confusion
				id = -id
			}
			
			existing_rows, err_row := sel.Query(id)
			if err_row != nil {
				fmt.Println("Select Error:", err)
				return
			}
			
			found:=false
			for existing_rows.Next() {
				found=true
			}
			
			if !found {
				// Only insert a new row if the row was not found
				fmt.Printf("Inserting row %d\n", id)
				
				_, err = ins.Exec(id, steps, 0)
				if err != nil {
					fmt.Println("Insert Error:", err)
					return
				}
			}
		}
	}
	
}

// training examples are stored in the dbs with negative ids.  All ids output from here are corrected to be positive
func list_of_interesting_problems_from_db(steps int, count int, is_training bool) []int {
	db := get_db_connection()
	defer db.Close()
	
	problem_list := []int{}
	
	filter_training_or_test := "id>0"
	if is_training {
		filter_training_or_test = "id<0"
	}
	
	// Only actually need id back
	rows, err := db.Query("SELECT id, steps, solution_count, currently_processing FROM problems"+
							" WHERE currently_processing=0 AND steps=? AND "+filter_training_or_test+
							" ORDER BY solution_count ASC"+ 
							" LIMIT ?", 
							steps, count)
	if err != nil {
		fmt.Println("Query interesting problems Error:", err)
		return problem_list
	}

	update, err := db.Prepare("UPDATE problems SET currently_processing=1 WHERE id=?")
	if err != nil {
		fmt.Println("Update 'currently_processing' Prepare Error:", err)
		return problem_list
	}
	defer update.Close()
	
	for rows.Next() {
		var id, steps, solution_count,currently_processing int
		err = rows.Scan(&id, &steps, &solution_count, &currently_processing)
		//err = rows.Err() // get any error encountered during iteration
		if err != nil {
			fmt.Println("Query interesting problem Error:", err)
			return problem_list
		}
		
		_, err = update.Exec(id)
		if err != nil {
			fmt.Println("Update 'currently_processing' Exec Error:", err)
			return problem_list
		}

		if is_training {
			id=-id
		}
		problem_list = append(problem_list, id)
	}
	
	return problem_list
}

func reset_all_currently_processing(is_training bool) { 
	db := get_db_connection()
	defer db.Close()
	
	filter_training_or_test := "id>0"
	if is_training {
		filter_training_or_test = "id<0"
	}
	
	_, err := db.Exec("UPDATE problems SET currently_processing=0 WHERE "+filter_training_or_test)
	if err != nil {
		fmt.Println("Reset currently_processing Error:", err)
		return 
	}
}

func get_unprocessed_seed_from_db(id int, is_training bool) int {
	db := get_db_connection()
	defer db.Close()
	
	if is_training {
		id = -id
	}

	seed:=1
	rows, err := db.Query("SELECT MAX(seed)+1 FROM solutions WHERE id=?", id)
	if err != nil {
		fmt.Println("Query seed_max Error:", err)
		return seed
	}

	for rows.Next() {
		var seed_max sql.NullInt64
		err = rows.Scan(&seed_max)
		if err != nil {
			fmt.Println("Query seed_max row Error:", err)
			return seed
		}
		if seed_max.Valid {
			seed = int(seed_max.Int64)
		}
	}
	return seed
}


func save_solution_to_db(id int, steps int, seed int, individual_result *IndividualResult, is_training bool) {
	// add to solutions
	db := get_db_connection()
	defer db.Close()
	
	if is_training {
		id = -id // Fix it up
	}
	
	// insert into the solutions db
	_, err := db.Exec("INSERT INTO solutions SET id=?, steps=?, seed=?, version=?, iter=?,"+
						" ones_i=?, mtsi=?, mtei=?,"+
						" ones_f=?, mtsf=?, mtef=?,"+
						" start=?",
						id, steps, seed,
						currently_running_version, 
						individual_result.iter, 
						individual_result.true_start_1s, 
						individual_result.mismatch_from_true_start_initial, 
						individual_result.mismatch_from_true_end_initial, 
						individual_result.true_end_1s, 
						individual_result.mismatch_from_true_start_final, 
						individual_result.mismatch_from_true_end_final, 
						individual_result.individual.start.toCompactString(),
					)
	if err != nil {
		fmt.Println("Inserting into solutions table for individual Error:", err)
		return
	}
	
	// increment problems solution_count, and reset currently_processing
	_, err = db.Exec("UPDATE problems SET solution_count=solution_count+1, currently_processing=0 WHERE id=?", id)
	if err != nil {
		fmt.Println("Updating problems table for individual Error:", err)
		return 
	}
}

// only_submit_for_steps_equals : Set this for +ve to filter submission to include only specific steps answers (rest are zeroed as a base-line)
func create_submission(fname string, is_training bool, only_submit_for_steps_equals int) {
	id_list := []int{}
	
	if is_training { // false for real submission, true for testing vs training_fake data
		for i:=-60001; i>=-61000; i-- { // Careful of the signs!
			id_list = append(id_list, i)
		}
		//id_list = append(id_list, -54)
	} else {  // THIS IS THE REAL DEAL!!
		for i:=1; i<=50000; i++ {
			id_list = append(id_list, i)
		}
	}
	
	db := get_db_connection()
	defer db.Close()
	
	file, err := os.Create(fname)
	if err != nil {
		fmt.Println("File Creation Error:", err)
		return
	}
	defer file.Close()
	
	file.WriteString("id")
	for i:=1; i<=400; i++ {
		file.WriteString(fmt.Sprintf(",start.%d", i))
	}
	file.WriteString("\n")
	
	query, err := db.Prepare("SELECT steps,iter,seed,version,mtei,mtef,start FROM solutions WHERE id=?")
	if err != nil {
		fmt.Println("Query solutions prepare Error:", err)
		return
	}
	defer query.Close()

	count_ids_found, count_zeroes_submitted := 0,0
	for _, id := range id_list {
		rows, err := query.Query(id)
		if err != nil {
			fmt.Println("Query solutions row for id=%d Error:", err)
			return
		}

		type BestRow struct {
			start_board *Board_BoolPacked
			steps,iter,seed,version int
			mtei, mtef int
			valid bool
		}
		best := BestRow{valid:false}
		id_found:=false
		
		samples_found:=0
		
		submit_zero_for_this_id:=false
		stats := NewBoardStats(board_width, board_height)
		for rows.Next() {
			var start string
			var steps,iter,seed,version int
			var mtei, mtef int
			err = rows.Scan(&steps, &iter, &seed, &version, &mtei, &mtef, &start)
			if err != nil {
				fmt.Println("Query start for id=%d Error:", id, err)
				return 
			}
			id_found = true

			// This zeroes out whole id if condition on one of it's rows fails
			if only_submit_for_steps_equals>0 && steps!=only_submit_for_steps_equals {
				submit_zero_for_this_id=true
				continue
			}
			
			// This ignores this particular row for this id
			// If it's the only row, then we'll end up with a zero for this id
			//   but that won't be included in the count_zeroes_submitted counter
			if version==999 { 
				continue
			}
			
			//if !(seed==1) {
			//if !(seed==1 || seed==2) {
			//if !(seed==1 || seed==2 || seed==3) {
			//if !(version==1002) {
			
			//if !(seed==4) {
			//if !(seed==4 || seed==5) {
			if !(version==1016) {
				
			//if !(seed==1 || seed==2 || seed==3 || seed==4) {
			//if !(seed==7) {
			//if !(seed==7 || seed==8) {
			//if !(version==1018) {
			//if !(seed==4 || seed==7) { // 1016x1 + 1018x1
				continue
			}
			
			start_board := NewBoard_BoolPacked(board_width, board_height)
			start_board.fromCompactString(start)
			
			// Do every entry twice ( so that an additional +1 for the best will tie-break a 50/50 threshold)
			start_board.AddToStats(stats)
			start_board.AddToStats(stats)
			//fmt.Println(start_board)
			samples_found++
			
			// Figure out whether this is going to be the best board for tie-breaking
			this_is_better_than_current_best := !best.valid // This picks up the first one immediately
			
			// If we ended up with fewer errors, or got to the same level, but quicker, this is probably better
			if mtef<best.mtef || (mtef<best.mtef && iter<best.iter) { 
				this_is_better_than_current_best=true
			}
			
			if this_is_better_than_current_best {
				best = BestRow{start_board, steps, iter, seed, version, mtei, mtef, true}
				fmt.Printf("Best @%5d now %v\n", id, best)
			}
		}

		threshold := 50
		
		
		// Now add a single instance of best board to the stats to act as a tie-breaker 
		if best.valid {
			best.start_board.AddToStats(stats)
			
			mtef_bar := 999999
			iter_bar := 999999
			
			if best.steps == 1 {
				//mtef_bar>15  // discredited
				//iter_bar>650 // discredited
				threshold = 50
				//threshold = 65 // This is worse than 50 for 1016
				//threshold = 85  // worse for 1002 and 1016
			}
			if best.steps == 2 {
				// if best.mtef>4 || best.iter>350 {  // iter criteria discredited on fake training
				// iter_bar=350   // iter criteria discredited on fake training
				//mtef_bar=20 // no effect
				threshold = 50
				//threshold = 65 // This is worse than 50 for 1016
			}
			if best.steps == 3 {
				// iter_bar=350   // iter criteria discredited on fake training
				mtef_bar = 10
				//threshold = 50
				threshold = 65
				//threshold = 85
			}
			if best.steps == 4 {
				iter_bar=350  
				mtef_bar=5
				threshold = 50
				threshold = 65
				//threshold = 85
			}
			if best.steps == 5 {
				// iter_bar=350   // iter criteria discredited on fake training
				mtef_bar=10
				//threshold = 30
				threshold = 50 // also works for 1016 (but worse for 1002)
				threshold = 65 
				//threshold = 75 
				//threshold = 85 // flexible for 1016 and/or 1002
			}
			if true {
				if best.iter > iter_bar || best.mtef> mtef_bar {
					//submit_zero_for_this_id = true
				}
			}
			/*
			threshold=50
			if samples_found==2 {
				threshold=65 // Better for 5s than 50
			}
			if samples_found==3 {
				threshold=65 // Better for 5s than 50
			}
			*/
		}
		if id_found {
			count_ids_found++
		}
		
		//fmt.Println(stats)
		
		// Ok, so now let's figure out a board from these stats that's a better guess
		guess_board := NewBoard_BoolPacked(board_width, board_height)
		guess_board.ThresholdStats(stats, threshold)
		
		//guess_board.ThresholdStats(stats, 50) // This needs 1/2 or 2/3 or 2/4 i.e. a majority or more
		//guess_board.ThresholdStats(stats, 65) // This needs 2/2 or 2/3 or 3/4
		//guess_board.ThresholdStats(stats, 85) // This needs 2/2 or 3/3 or 4/4 i.e. all
		
		//fmt.Println(guess_board)
		
		// This implements a filter
		if submit_zero_for_this_id {
			// Instead of the true guess, zero it all out
			guess_board = NewBoard_BoolPacked(board_width, board_height)
			count_zeroes_submitted++
		}
		
		id_positive:=id
		if is_training { // This is for CSV-land so all ids are positive
			id_positive=-id
		}
		file.WriteString(fmt.Sprintf("%d", id_positive))
		file.WriteString(guess_board.toCSV())
		file.WriteString("\n")
	}
	if count_ids_found == len(id_list) {
		if count_ids_found==50000 {
			if count_zeroes_submitted>0 {
				non_zero := count_ids_found-count_zeroes_submitted
				fmt.Printf("FILTERED SUBMISSION FILE :: ONLY HAS %d delta=%d non-zero entries\n", non_zero, only_submit_for_steps_equals) 
			}
			fmt.Printf("TODO : gzip %s\n", fname) 
		} else {
			fmt.Printf("Test file created : %s\n", fname) 
		}
	} else {
		fmt.Printf("BAD SUBMISSION FILE :: ONLY HAS %d of %d IDs!\n", count_ids_found, len(id_list)) 
	}
}


/* :: Useful SQL ::
 * select id,steps,iter from solutions where id>0 and id<50 order by steps,id,iter
 * select steps,seed,count(id) from solutions where id>0 group by steps, seed
 * 
 * How many of each type do we have ::
 * select steps,count(id) from solutions where id>0 group by steps
 * 
 * Diagnose and Fix early aborted runs :: 
 * select steps,count(id) from problems where currently_processing>0 and id>0 group by steps
 * update problems set currently_processing=0 where id>0
 * 
 * What is the state of all solutions ::
 * select steps,solution_count,count(id) from problems where id>0 group by steps, solution_count order by steps,solution_count
 * 
 * Where are there holes to fill ::
 * select steps,solution_count,count(id) from problems where id>0 and solution_count=0 group by steps, solution_count order by steps
 * 
 * What is the distribution of solutions
 * select steps,solution_count,count(id) from problems where id>0 group by steps, solution_count order by steps, solution_count
 * select steps,seed,count(id) from solutions where id>0 group by steps, seed order by steps, seed
 * select steps,version,seed, count(id) from solutions where id>0 group by steps, version, seed order by steps, version, seed
 * 
 * What is currently being worked on ::
 * select steps,solution_count,count(id) from problems where id>0 and currently_processing=1 group by steps, solution_count order by steps
 * 
 * Clean out the solutions a little
 * delete from solutions where id>0 and seed=4
 * update problems set solution_count=3 where id>0 and solution_count=4 
 */
