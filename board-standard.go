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
	"sort"
)

import (
	"image"
	"image/color"
	"image/draw"
	"image/png"
)


/****************************************************************************************/
// The following will work on more generalized GoL mechanics, but are 10x slower
/****************************************************************************************/

func (f *Board_BoolPacked) isSet_safe(x, y int) bool {
	if x<0 || x>=board_width || y<0 || y>=board_height {
		return false
	}
	return f.isSet(x,y)
}

func (f *Board_BoolPacked) Set_safe(x, y int, b bool) {
	if x<0 || x>=board_width || y<0 || y>=board_height {
		return 
	}
	f.Set(x,y,b)
}


// puts Board in a random state
func (f *Board_BoolPacked) UniformRandom(pct float32) {
	for y := 0; y < f.h; y++ {
		for x := 0; x < f.w; x++ {
			f.Set(x, y, (rand.Float32() < pct))
		}
	}
}

// loads Board from a string : Using '\n' and 'X' as markers
func (f *Board_BoolPacked) LoadString(s string) {
	x := 0
	y := 0
	for _, v := range s[:] {
		if v == 'X' {
			f.Set(x, y, true)
		}
		x++
		if v == '\n' {
			x = 0
			y++
		}
	}
}

func (f *Board_BoolPacked) LoadArray(csv_strings []string) {
	x := 0
	y := 0

	for _, v := range csv_strings[:] {
		if v == "1" {
			f.Set(x, y, true)
			//fmt.Print("*")
		}
		x++
		if x >= f.w {
			x = 0
			y++
		}
	}
}

// Next returns the state of the specified cell at the next time step.
func (f *Board_BoolPacked) IterateCell(x, y int) bool {
	// Count the adjacent cells that are alive.
	alive := 0
	for i := -1; i <= 1; i++ {
		for j := -1; j <= 1; j++ {
			if (j != 0 || i != 0) && f.isSet(x+i, y+j) {
				alive++
			}
		}
	}
	// Return next state according to the game rules:
	//   exactly 3 neighbors: on,
	//   exactly 2 neighbors: maintain current state,
	//   otherwise: off.
	return alive == 3 || alive == 2 && f.isSet(x, y)
}

func (f *Board_BoolPacked) Iterate_Generic(next *Board_BoolPacked) {
	// Update the state of the next field (next) in-place from the current field (f).
	for y := 0; y < f.h; y++ {
		for x := 0; x < f.w; x++ {
			next.Set(x, y, f.IterateCell(x, y))
		}
	}
}

// String returns the game board as a string.
func (f *Board_BoolPacked) String() string {
	var buf bytes.Buffer
	outer := 1
	for y := 0 - outer; y < f.h+outer; y++ {
		for x := 0 - outer; x < f.w+outer; x++ {
			b := byte('-')
			if x < 0 || x >= f.w || y < 0 || y >= f.h {
				b = '0'
			} else { 
				if f.isSet(x, y) {
					b = '*'
				}
			}
			buf.WriteByte(b)
		}
		buf.WriteByte('\n')
	}
	return buf.String()
}

// Returns the game board as a string of 1s and 0s with commas (with a preceeding ',')
func (f *Board_BoolPacked) toCSV() string {
	var buf bytes.Buffer
	for y := 0; y < f.h; y++ {
		for x := 0; x < f.w; x++ {
			b := byte('0')
			if f.isSet(x, y) {
				b = '1'
			}
			buf.WriteByte(',')
			buf.WriteByte(b)
		}
	}
	return buf.String()
}

// Returns the game board as a string of 1s and 0s (compact)
func (f *Board_BoolPacked) toCompactString() string {
	var buf bytes.Buffer
	for y := 0; y < f.h; y++ {
		for x := 0; x < f.w; x++ {
			b := byte('0')
			if f.isSet(x, y) {
				b = '1'
			}
			buf.WriteByte(b)
		}
	}
	return buf.String()
}
// Reads the game board as a string of 1s and 0s (compact)
func (f *Board_BoolPacked) fromCompactString(buf string)  {
	for y,i := 0,0; y < f.h; y++ {
		for x := 0; x < f.w; x++ {
			f.Set(x,y, buf[i] == '1')
			i++
		}
	}
}


func (f *Board_BoolPacked) AddToStats(bs *BoardStats) {
	for y := 0; y < f.h; y++ {
		for x := 0; x < f.w; x++ {
			if f.isSet(x, y) {
				bs.freq[y][x]++
			}
		}
	}
	bs.count++
}

func (f *Board_BoolPacked) ThresholdStats(bs *BoardStats, threshold_level_pct int) {
	for y := 0; y < f.h; y++ {
		for x := 0; x < f.w; x++ {
			f.Set(x, y, bs.freq[y][x]*100>threshold_level_pct*bs.count)
		}
	}
}

type BoardStats struct {
	freq  [][]int
	w, h  int
	count int
	mismatch_amount int
}

// NewField_BoolArray returns an empty field of the specified width and height.
func NewBoardStats(w, h int) *BoardStats {
	freq := make([][]int, h)
	for i := range freq {
		freq[i] = make([]int, w)
	}
	//fmt.Print("CreatedBoardStats\n")
	return &BoardStats{freq: freq, w: w, h: h, count: 0, mismatch_amount:0}
}

func (bs *BoardStats) MisMatchBy(mismatch int) {
	bs.mismatch_amount = mismatch
}

// BoardIterator stores the state of a round of Conway's Game of Life.
type BoardIterator struct {
	current, temp_internal_only *Board_BoolPacked
}

// BoardIterator returns a new Life game state
func NewBoardIterator(w, h int) *BoardIterator {
	return &BoardIterator{
		current: NewBoard_BoolPacked(w, h), 
		temp_internal_only: NewBoard_BoolPacked(w, h),
	}
}

// Step advances the game by one instant, recomputing and updating all cells.
func (bi *BoardIterator) Iterate(n int) {
	for i := 0; i < n; i++ {
		bi.current.Iterate(bi.temp_internal_only)
		// Now swap boards, to put the result in prime position
		bi.current, bi.temp_internal_only = bi.temp_internal_only, bi.current
	}
}

type LifeProblem struct {
	id         int
	start, end *Board_BoolPacked
	steps      int
	// Finished, iterations, confidence, etc
}

func (problem *LifeProblem) CreateFake() {
	id := problem.id
	steps := problem.steps
	
	start := NewBoard_BoolPacked(board_width, board_height)
	end   := NewBoard_BoolPacked(board_width, board_height)
	
	empty := NewBoard_BoolPacked(board_width, board_height)
	for found:=false; !found; {
		// create a board with a random initial density U(0..1) 
		uniform := rand.Float32()
		fmt.Printf("id[%6d].steps=%d, Uniform Density = %6.4f\n", id, steps, uniform)
		
		initial := NewBoard_BoolPacked(board_width, board_height)
		initial.UniformRandom(uniform)
		//fmt.Println(initial)
		
		// transition it forwards 5 times
		l := NewBoardIterator(board_width, board_height)
		l.current.CopyFrom(initial)
		
		l.Iterate(5)
		
		// Now l.current is the actual start board
		start.CopyFrom(l.current) // this overwrites...
		//fmt.Println(start)
		
		// iterate forward the appropriate number of steps
		l.Iterate(steps)
		
		// Now l.current is the actual ending board
		end.CopyFrom(l.current)
		//fmt.Println(end)
		
		// if end is not empty, then we've succeeded
		if end.CompareTo(empty, nil) > 0 {
			found = true
			//fmt.Println("Success!")
		}
	}
	
	problem.start = start
	problem.end = end
}

type LifeProblemSet struct {
	problem map[int]LifeProblem
	is_training bool
	
	transition_collection []TransitionCollectionList
}

// Unlike the db, the ids here match the training.csv and test.csv files exactly
func (s *LifeProblemSet) load_csv(is_training bool, id_list []int) {
	filename := "data/test.csv"
	if is_training {
		filename = "data/train.csv"
		if id_list[0]>50000 { 
			filename = "data/train_fake.csv"
		}
	}
	s.load_csv_from_file(filename, is_training, id_list)
}

// Unlike the db, the ids here match the training.csv and test.csv files exactly
func (s *LifeProblemSet) load_csv_from_file(filename string, is_training bool, id_list []int) {
	if s.problem == nil {
		s.problem = make(map[int]LifeProblem)
	}
	s.is_training = is_training
	
	file, err := os.Open(filename)
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

	id_max := 0
	id_map := make(map[int]bool)
	for _, id := range id_list {
		id_map[id] = true
		if id_max < id {
			id_max = id
		}
	}

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
		if id_map[id] {
			//fmt.Println(record) // record has the type []string
			steps, _ := strconv.Atoi(record[1])

			start := NewBoard_BoolPacked(board_width, board_height)
			end := NewBoard_BoolPacked(board_width, board_height)
			if is_training {
				start.LoadArray(record[2:402])
				end.LoadArray(record[402:802])
			} else {
				end.LoadArray(record[2:402])
			}

			s.problem[id] = LifeProblem{
				id:    id,
				start: start,
				end:   end,
				steps: steps,
			}
			fmt.Printf("Loaded problem[%d] : steps=%d\n", id, steps)
			//fmt.Print(s.problem[id].start)
		}
		if id > id_max {
			return // fact-of-life : ids are ascending order, so can quit reading early
		}
	}
}

// Unlike the db, the ids here match the csv files exactly
func (s *LifeProblemSet) save_csv(filename string) { // = "data/train_fake.csv"
	file, err := os.Create(filename)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	defer file.Close()

	// Header
	file.WriteString("id,delta")
	for i:=1; i<=400; i++ {
		file.WriteString(fmt.Sprintf(",start.%d", i))
	}
	for i:=1; i<=400; i++ {
		file.WriteString(fmt.Sprintf(",stop.%d", i))
	}
	file.WriteString("\n")

	// This file should be in id-ascending order
	ids := []int{}
	for _, problem := range s.problem {
		ids = append(ids, problem.id)
	}
	sort.Ints(ids)

	for _, id := range ids {
		problem := s.problem[id]
		file.WriteString(fmt.Sprintf("%d,%d", problem.id, problem.steps))
		
		file.WriteString(problem.start.toCSV())
		file.WriteString(problem.end.toCSV())
		file.WriteString("\n")
	}
}

func determine_kaggle_score(fake_training_data_csv string, submission_csv string) float32 {
	var training_data, submission LifeProblemSet
	id_list := []int{}
	for i:=-60001; i<=61000; i-- {
		id_list = append(id_list, i)
	}
	
	training_data.load_csv_from_file(fake_training_data_csv, true, id_list)
	submission.load_csv_from_file(submission_csv, true, id_list)

	total_errors := 0
	total_boards := 0
	for _,id := range id_list {
		total_errors += training_data.problem[id].start.CompareTo(submission.problem[id].start, nil)
		total_boards++
	}
	score := float32(total_errors)/float32(total_boards)
	return score
}

func (s *LifeProblemSet) load_transition_collection(steps int) {
	// Only load if it's not already loaded
	if s.transition_collection == nil {
		s.transition_collection = make([]TransitionCollectionList, 10) // Much too long (no worries)
	}
	if len(s.transition_collection[steps].pre)==0 {
		//fmt.Printf("******** STEPS OVERRIDE ************\n")
		//s.transition_collection[steps].LoadCSV(fmt.Sprintf(TransitionCollectionFileStrFmt, 1)) 
		s.transition_collection[steps].LoadCSV(fmt.Sprintf(TransitionCollectionFileStrFmt, steps)) 
	}
}


type ImageSet struct {
	im                       *image.RGBA
	rows, cols               int
	row_current, col_current int
}

func NewImageSet(rows, cols int) *ImageSet {
	im := image.NewRGBA(image.Rect(0, 0, cols*(board_width+2)+2, rows*(board_height+2)+2))                             //*NRGBA (image.Image interface)
	draw.Draw(im, im.Bounds(), image.NewUniform(color.RGBA{98, 166, 255, 255}), image.ZP, draw.Src) // color.Transparent
	return &ImageSet{
		im:   im,
		rows: rows, cols: cols,
		row_current: 0, col_current: 0,
	}
}

func (i *ImageSet) save(f string) {
	w, _ := os.Create(f)
	defer w.Close()
	png.Encode(w, i.im)
}

func (i *ImageSet) DrawStats(row, col int, bs *BoardStats) {
	offset_x := col*(board_width+2) + 2
	offset_y := row*(board_height+2) + 2

	for x := 0; x < bs.w; x++ {
		for y := 0; y < bs.h; y++ {
			g := bs.freq[y][x] * 255 / bs.count
			if bs.mismatch_amount>0 {
				pct := 100 - bs.mismatch_amount * 50 / 100
				if pct<0 {
					pct=0
				}
				//fmt.Printf("Mismatch pct=%d\n", pct)
				g = (g*pct) /100
			}
			i.im.Set(offset_x+x, offset_y+y, color.Gray{uint8(g)})
		}
	}
}

func (i *ImageSet) DrawStatsNext(bs *BoardStats) {
	i.DrawStats(i.row_current, i.col_current, bs)
	i.col_current++
	if i.col_current >= i.cols {
		i.DrawStatsCRLF()
	}
}

func (i *ImageSet) DrawStatsCRLF() {
	i.col_current = 0
	i.row_current++
	if i.row_current >= i.rows {
		//fmt.Print("New Beginning\n")
		i.row_current = 0
	}
}



