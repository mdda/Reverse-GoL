// An implementation of Conway's Game of Life.
// See reverse-gol.go for build/run

package main

import (
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"strconv"
)


type TransitionCollection struct {
	pre map[Patch]map[Patch]bool
}

func (t *TransitionCollection) training_csv_to_stats(f string, step_filter int) {
	if t.pre == nil {
		t.pre = make(map[Patch]map[Patch]bool)
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

	record_count:=0
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
		if steps==step_filter && id>100 && id<200000 { // Ignore first few records (for now)
			//fmt.Println(record) // record has the type []string
			//fmt.Printf("id[%5d].steps=%d\n", id, steps) 
			
			start := NewBoard_BoolPacked(board_width, board_height)
			end   := NewBoard_BoolPacked(board_width, board_height)
			
			start.LoadArray(record[2:402])
			end.LoadArray(record[402:802])
			
			//fmt.Println(end)
			existing_map_count:=0
			
			// Go through the end and start boards in lock-step
			// Create ASAP the mapping end->start
			for y:=0; y<end.h; y++ {
				for x:=0; x<end.w; x++ {
					p:=start.MakePatch(x,y)
					q:=end.MakePatch(x,y)
					if len(t.pre[q])>0 {
						// If end_rep exists : no need to create map, it's already there
						//fmt.Printf("id[%5d]@(%2d,%2d) exists in map (prior len:%5d) = %25b\n", id, x ,y, len(t.pre[q]), q)
						existing_map_count++
					} else {
						// If end_rep does not exist : Make [start_rep] the array dangling off end_rep
						//fmt.Printf("id[%5d]@(%2d,%2d) creating fresh map for %25b!\n", id, x,y, q)
						t.pre[q] = make(map[Patch]bool)
					}
					// Add start_rep to the map dangling off end_rep
					t.pre[q][p]=true
				}
			}
			fmt.Printf("id[%5d].steps=%d - existing=%3d/400\n", id, steps, existing_map_count) 
			record_count++
		}
	}
	fmt.Printf("Total end-map count : %7d\n", len(t.pre)) 
	fmt.Printf("Total record  count : %7d\n", record_count) 
}

