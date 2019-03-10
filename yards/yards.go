package main

import (
	"encoding/csv"
	"fmt"
	"io"
	"log"
	"math"
	"os"
	"strconv"
)

func points(pts int) string {
	if pts == 0 {
		return "0"
	}
	if pts <= 6 {
		return "1-6"
	}
	if pts <= 13 {
		return "7-13"
	}
	if pts <= 20 {
		return "14-20"
	}
	if pts <= 27 {
		return "21-27"
	}
	if pts <= 34 {
		return "28-34"
	}
	return "35+"
}

func yards(yds int) int {
	if yds < 0 {
		return 10
	}
	if yds < 100 {
		return 7
	}
	if yds < 200 {
		return 4
	}
	if yds < 300 {
		return 1
	}
	if yds < 400 {
		return 0
	}
	if yds < 500 {
		return -1
	}
	return -4
}

func main() {
	m := map[string]map[int]int{
		"0":     make(map[int]int),
		"1-6":   make(map[int]int),
		"7-13":  make(map[int]int),
		"14-20": make(map[int]int),
		"21-27": make(map[int]int),
		"28-34": make(map[int]int),
		"35+":   make(map[int]int),
	}
	f, err := os.Open("games-2017.csv")
	if err != nil {
		log.Fatal(err)
	}
	r := csv.NewReader(f)
	for {
		record, err := r.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatal(err)
		}
		pts1, err := strconv.Atoi(record[8])
		if err != nil {
			log.Fatal(err)
		}
		pts2, err := strconv.Atoi(record[9])
		if err != nil {
			log.Fatal(err)
		}
		yds1, err := strconv.Atoi(record[10])
		if err != nil {
			log.Fatal(err)
		}
		yds2, err := strconv.Atoi(record[12])
		if err != nil {
			log.Fatal(err)
		}
		m[points(pts1)][yards(yds1)]++
		m[points(pts2)][yards(yds2)]++
	}
	for _, p := range []string{"0", "1-6", "7-13", "14-20", "21-27", "28-34", "35+"} {
		hist := m[p]
		total := 0
		count := 0
		for y, n := range hist {
			total += n * y
			count += n
		}
		f := float64(total) / float64(count)
		r := int(math.Floor(f + 0.5))
		fmt.Printf("%5s %+2d %+5.2f\n", p, r, f)
	}
}
