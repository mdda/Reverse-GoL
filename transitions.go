// An implementation of Conway's Game of Life.
// See reverse-gol.go for build/run

package main

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"io"
	"math/rand"
	"os"
	"strconv"
)


type TransitionCollection struct {
	pre map[int][]int
}

func (t *TransitionCollection) training_csv_to_stats(f string, step_filter int) {
	if t.pre == nil {
		t.pre = make(map[int][]int)
	}
	file, err := os.Open(f)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	defer file.Close()
	reader := csv.NewReader(file)

	// First line different
	header, err := reader.Read()
	if header[0] != "id" {
		fmt.Println("Bad Header", err)
		return
	}
	//fmt.Println("Header Start: ", header[2:402])
	//fmt.Println("Header Stop : ", header[402:802])

	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		} else if err != nil {
			fmt.Println("Error:", err)
			return
		}

		// record is []string
		id, _ := strconv.Atoi(record[0])
		steps, _ := strconv.Atoi(record[1])
		if steps==step_filter && id>100 { // Ignore first few records (for now)
			//fmt.Println(record) // record has the type []string
			
			start := NewBoard_BoolPacked(board_width, board_height)
			end   := NewBoard_BoolPacked(board_width, board_height)
			
			start.LoadArray(record[2:402])
			end.LoadArray(record[402:802])
			
			// Go through the end and start boards in lock-step
			// Create ASAP the mapping end->start
			// Use isSet_safe()
			
			// If end_rep exists : Add start_rep to the array dangling off end_rep
			// If end_rep !exists : Make [start_rep] the array dangling off end_rep
		}
	}
}

