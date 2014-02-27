// An implementation of Conway's Game of Life.
// See reverse-gol.go for build/run

package main

import (
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"strconv"
	"bytes"
	"math/rand"
	"sort"
)


type Patch int

func (f *Board_BoolPacked) MakePatch(x,y int) Patch {
	var p Patch=0
	for dy:=-2; dy<=+2; dy++ {
		for dx:=-2; dx<=+2; dx++ {
			p<<=1
			if f.isSet_safe(x+dx,y+dy) {
				p |= 1
			}
		}
	}
	return p
}

func (f *Board_BoolPacked) OverlayPatch(x,y int, p Patch) {
	//fmt.Printf("Overlaying Patch @(%d,%d)\n", x,y)
	//fmt.Print(p)
	//fmt.Print(f)
	// Do the creation process, but in reverse
	for dy:=+2; dy>=-2; dy-- {
		for dx:=+2; dx>=-2; dx-- {
			f.Set_safe(x+dx, y+dy, (p & 1) !=0) // bit-twiddle
			p>>=1
		}
	}
	//fmt.Print(f)
}

func (p Patch) isSet(x,y int) bool {
	return p & (1<<uint((4-x)+5*(4-y))) !=0
}

func (p Patch) String() string {
	var buf bytes.Buffer
	outer := 0
	for y := 0 - outer; y < 5+outer; y++ {
		for x := 0 - outer; x < 5+outer; x++ {
			b := byte('-')
			if x < 0 || x >= 5 || y < 0 || y >= 5 {
				b = '?'
			} else { 
				if p.isSet(x,y) {
					b = '*'
				}
			}
			buf.WriteByte(b)
		}
		buf.WriteByte('\n')
	}
	return buf.String()
}

func (p Patch) Flip_UD() Patch {
	temp:=int(p)
	var q int=0
	
	// This is a boolean operation - batches of 5...
	j := (1<<4 | 1<<3 | 1<<2 | 1<<1 | 1<<0) // 0b 00000 00000 00000 00000 11111
	for i:=0; i<5; i++ {
		q <<= 5 // Shift existing stuff 'to the left'
		q |= (temp & j) //  Next row into empty space
		temp >>= 5 // Shift next block into poll position
	}
	return Patch(q)
}

func (p Patch) Flip_LR() Patch {
	temp:=int(p)
	var q int=0
	
	// This is a boolean operation - pass a comb over data
	j   := (1<<20 | 1<<15 | 1<<10 | 1<<5 | 1<<0) // 0b 00001 00001 00001 00001 00001
	
	for y:=0; y<5; y++ {
		q <<= 1 // Shift existing stuff 'to the left'
		q |= (temp & j) //  plop columns into empty space(s)
		temp >>= 1 // Shift next columns into poll position(s)
	}
	return Patch(q)
}

type PatchOrientation struct {
	patch Patch
	flip_ud, flip_lr bool
}

func (p Patch) BestOrientation() PatchOrientation {  
	var orientation [4]PatchOrientation
	
	// Score the different possibilities
	orientation[0] = PatchOrientation{ p, false, false}
	orientation[1] = PatchOrientation{ p.Flip_UD(), true, false }
	orientation[2] = PatchOrientation{ orientation[0].patch.Flip_LR(), false, true }
	orientation[3] = PatchOrientation{ orientation[1].patch.Flip_LR(), true, true }
	
	best_orientation := orientation[0]
	for _,this_orientation := range orientation {
		//fmt.Printf("Patch[%d]=%8d %d %d\n", i, int(this_orientation.patch), this_orientation.flip_ud, this_orientation.flip_lr)
		if this_orientation.patch < best_orientation.patch {
			best_orientation = this_orientation
		}
	}
	//fmt.Printf("Best Orientation : (%d,%d) is %8d\n", best_orientation.flip_ud, best_orientation.flip_lr, int(best_orientation.patch)) 
	return best_orientation
}

func (tc *TransitionCollectionList) GetRandomEntry_OrientationCompensated(q Patch) Patch {
	oriented := q.BestOrientation()
	
	if pl, ok :=tc.pre[oriented.patch]; ok {
		// if found, then copy a random one of its starters into the new individual
		//fmt.Printf("Found known end!\n")
		p := pl.GetRandomEntry()
		
		// Do the same (best) orientation maneuver on p
		if oriented.flip_ud {
			p = p.Flip_UD()
		}
		if oriented.flip_lr {
			p = p.Flip_LR()
		}
		
		return p
	}
	//fmt.Printf("Did not find known end!\n")
	
	return Patch(-1)
}

type PatchFreq struct {
	patch Patch
	freq int
}
type PatchMap  map[Patch]int
type PatchList struct {
	starts []PatchFreq
	freq_total int
}

/*
func (starts PatchMap) GetRandomEntryUNUSED() Patch {  // Not in use because of problems below
	n_starts := len(starts)

	if false {
		// This is awkward because this is a map, and have to iterate through to find random entry
		// BUT :: This has a problem : The order of accesses from the map is 'quasi random' 
		//        - but not due to "math/rand" :=> It's non-deterministic.  Which sucks
		start_random_index := rand.Intn(n_starts)
		for k,_ := range starts {
			start_random_index--
			if start_random_index<0 {
				return k  //  Return immediately - have iterated though to right (random) place
			}
		}
		fmt.Printf("Unable to find a random start for known end\n")
		return Patch(-1)
	}
	
	if n_starts==1 {
		for k,_ := range starts {
			return k //  Return immediately - only one choice, after all
		}
	}
	
	// Only way to get deterministic behaviour is to sort the list of keys
	// and pick the 'start_random_index'th one
	start_list := make([]int, n_starts)
	i:=0
	for k,_ := range starts {
		start_list[i] = int(k)
		i++
	}
	
	sort.Ints(start_list)
	
	start_random_index := rand.Intn(n_starts)
	return Patch(start_list[start_random_index])
}
*/

func (pl PatchList) GetRandomEntry() Patch {
	n_starts := len(pl.starts)
	start_random_index := rand.Intn(n_starts)
	return pl.starts[start_random_index].patch
}

type TransitionCollectionMap struct {
	pre map[Patch]PatchMap
}

type TransitionCollectionList struct {
	pre map[Patch]PatchList
}

const TransitionCollectionFileStrFmt = "stats/transition-%d.csv"

func (t *TransitionCollectionMap) TrainingCSV_to_stats(f string, step_filter int) {
	if t.pre == nil {
		t.pre = make(map[Patch]PatchMap)
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
			
			existing_map_count:=0
			
			// Go through the end and start boards in lock-step
			// Create ASAP the mapping end->start
			for y:=0; y<end.h; y++ {
				for x:=0; x<end.w; x++ {
					p := start.MakePatch(x,y)
					q := end.MakePatch(x,y)
					
					if false {
						q_test, q_orig := q, q
						
						q_test = q_test.Flip_LR()
						q_test = q_test.Flip_UD()
						q_test = q_test.Flip_LR()
						q_test = q_test.Flip_UD()
						
						if q_test != q_orig {
							fmt.Printf("PATCH FLIP WHOOPS!\n") 
							fmt.Println(q_orig) 
							fmt.Println(q_test) 
						}
					}
					
					oriented := q.BestOrientation()
					q = oriented.patch
					
					// Do the same (best) orientation maneuver on p
					if oriented.flip_ud {
						p = p.Flip_UD()
					}
					if oriented.flip_lr {
						p = p.Flip_LR()
					}
					
					if len(t.pre[q])>0 {
						// If end_rep exists : no need to create map, it's already there
						//fmt.Printf("id[%5d]@(%2d,%2d) exists in map (prior len:%5d) = %25b\n", id, x ,y, len(t.pre[q]), q)
						existing_map_count++
					} else {
						// If end_rep does not exist : Make [start_rep] the array dangling off end_rep
						//fmt.Printf("id[%5d]@(%2d,%2d) creating fresh map for %25b!\n", id, x,y, q)
						t.pre[q] = make(PatchMap)
					}
					// Add start_rep to the map dangling off end_rep
					t.pre[q][p]++
					
					/*
					if q == 14336 { // This is '***'
						fmt.Printf("Precursor found %8d:\n", int(p))
						fmt.Println(p)  // This should be a vertical strip or '***' (depending on whether steps is odd or even)
					}
					*/
				}
			}
			fmt.Printf("id[%5d].steps=%d - existing=%3d/400\n", id, steps, existing_map_count) 
			if existing_map_count>=400*1000 {
				fmt.Println(end)
				return
			}
			record_count++
		}
	}
	fmt.Printf("Total end-map count : %7d\n", len(t.pre)) 
	fmt.Printf("Total record  count : %7d\n", record_count) 
}

type ByFreqDesc []PatchFreq
func (a ByFreqDesc) Len() int           { return len(a) }
func (a ByFreqDesc) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a ByFreqDesc) Less(i, j int) bool { return a[i].freq > a[j].freq }

func (t *TransitionCollectionMap) SaveCSV(f string) {
	file, err := os.Create(f)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	defer file.Close()
	
	// Need to determine total_freq, and potentially do sorting
	// starts is a map[Patch]int : Want to Create a PatchList from this PatchMap
	
	for end, starts := range t.pre {
		freq_total :=0
		
		// Allocate an array of the right size
		starts_list := make([]PatchFreq, len(starts))
		i:=0
		for start,freq := range starts {  
			starts_list[i] = PatchFreq{patch:start, freq:freq}
			freq_total += freq
			i++
		}
		
		// Sort the starts_list[] here...
		sort.Sort(ByFreqDesc(starts_list))
		
		// Write out the list as a CSV list 
		file.WriteString(fmt.Sprintf("%d,%d", int(end), freq_total))
		for _,start := range starts_list { 
			file.WriteString(fmt.Sprintf(",%d,%d", int(start.patch), start.freq))
		}
		file.WriteString("\n")
	}
}

func (t *TransitionCollectionList) LoadCSV(f string) {
	if t.pre == nil {
		t.pre = make(map[Patch]PatchList)
	}
	file, err := os.Open(f)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	defer file.Close()
	
	reader := csv.NewReader(file)
	reader.FieldsPerRecord = -1 // Allow for variable # of fields per line
	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		} else if err != nil {
			fmt.Println("Error:", err)
			return
		}

		// record is []string
		end, _ := strconv.Atoi(record[0])
		freq_total, _ := strconv.Atoi(record[1])
		starts := make([]PatchFreq, len(record)/2-1) // These are listed in {patch, freq} pairs
		for i,j:=2,0; i<len(record); i+=2 {
			//fmt.Printf("Loading col: %d\n", j)
			patch, _ := strconv.Atoi(record[i])
			freq, _  := strconv.Atoi(record[i+1])
			starts[j]=PatchFreq{patch:Patch(patch), freq:freq}
			j++
		}
		t.pre[Patch(end)] = PatchList{starts:starts, freq_total:freq_total}
	}
	fmt.Printf("Loaded %d transition end-points\n", len(t.pre))
}


