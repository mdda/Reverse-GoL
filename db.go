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
	`mtsi` int(11) DEFAULT NULL, 
	`mtei` int(11) NOT NULL, 
	`mtsf` int(11) DEFAULT NULL, 
	`mtef` int(11) NOT NULL, 
	`start` text NOT NULL, 
	KEY `solutions_id` (`id`) 
) ENGINE=InnoDB DEFAULT CHARSET=latin1
*/

func get_db_connection() *sql.DB {
	//db, err := sql.Open("mysql", "user:password@/database")
	db, err := sql.Open("mysql", "reverse-gol:reverse-gol@/reverse-gol")
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
	
	db.Exec("DELETE FROM PROBLEMS")
	
	ins, err := db.Prepare("INSERT INTO problems SET id=?, steps=?, solution_count=?")
	if err != nil {
		fmt.Println("Insert Prepare Error:", err)
		return
	}
	
	for i,f := range []string{"data/train.csv", "data/test.csv"} {
		fmt.Printf("Opening %s - %d\n", f, i)
		
		file, err := os.Open(f)
		if err != nil {
			fmt.Println("Error:", err)
			return
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
			
			if i==0 {
				// This is the training data : Let's give it a negative id to avoid confusion
				id = -id
			}
			_, err = ins.Exec(id, steps, 0)
			if err != nil {
				fmt.Println("Insert Error:", err)
				return
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
						" mtsi=?, mtei=?,"+
						" mtsf=?, mtef=?,"+
						" start=?",
						id, steps, seed,
						currently_running_version, 
						individual_result.iter, 
						individual_result.mismatch_from_true_start_initial, 
						individual_result.mismatch_from_true_end_initial, 
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

