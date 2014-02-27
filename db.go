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
	`solution_count` int(11) NOT NULL 
) ENGINE=InnoDB DEFAULT CHARSET=latin1
CREATE TABLE `solutions` ( 
	`id` int(11) NOT NULL, 
	`steps` int(11) NOT NULL, 
	`iter` int(11) NOT NULL, 
	`seed` int(11) NOT NULL, 
	`version` int(11) NOT NULL 
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

func list_of_interesting_problems_from_db(steps int, count int) []int {
	db := get_db_connection()
	defer db.Close()
	
	problem_list := []int{}
	
	//problem_list = append(problem_list, id)
	
	return problem_list
}

func add_individual_to_solutions_db(ind *Individual) {
	// add to solutions
	// increment problems solution_count
}
