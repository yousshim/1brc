package main

import (
	"bufio"
	"fmt"
	"io"
	"maps"
	"slices"
	"strconv"
)

type stationStat struct {
	min float64
	max float64
	sum float64
	cnt int
}

func (stat stationStat) String() string {
	mean := stat.sum / float64(stat.cnt)
	return fmt.Sprintf("%.1f/%.1f/%.1f", stat.min, stat.max, mean)
}

func Calculate(r io.Reader, w io.Writer) {
	s := bufio.NewScanner(r)
	stats := make(map[string]*stationStat)
	for s.Scan() {
		line := s.Bytes()
		splitIdx := slices.Index(line, ';')
		name := string(line[:splitIdx])
		temp, _ := strconv.ParseFloat(string(line[splitIdx+1:]), 64)
		if stat, ok := stats[name]; ok {
			stat.min = min(stat.min, temp)
			stat.max = max(stat.max, temp)
			stat.sum += temp
			stat.cnt++
		} else {
			stats[name] = &stationStat{
				min: temp,
				max: temp,
				sum: temp,
				cnt: 1,
			}
		}
	}
	for _, key := range slices.Sorted(maps.Keys(stats)) {
		fmt.Fprintf(w, "\"%s\"/%v\n", key, stats[key])
	}
}
