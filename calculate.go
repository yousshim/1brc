package main

import (
	"bufio"
	"fmt"
	"io"
	"maps"
	"slices"
)

type stationStat struct {
	min int
	max int
	sum int64
	cnt int
}

func (stat stationStat) String() string {
	mean := float64(stat.sum)/10/float64(stat.cnt)
	minn := float64(stat.min)/10
	maxx := float64(stat.max)/10
	return fmt.Sprintf("%.1f/%.1f/%.1f", minn, maxx, mean)
}

func Calculate(r io.Reader, w io.Writer) {
	s := bufio.NewScanner(r)
	stats := make(map[string]*stationStat)
	for s.Scan() {
		line := s.Bytes()
		splitIdx := slices.Index(line, ';')
		name := string(line[:splitIdx])
		temp := parseTemp(line[splitIdx+1:])
		if stat, ok := stats[name]; ok {
			stat.min = min(stat.min, temp)
			stat.max = max(stat.max, temp)
			stat.sum += int64(temp)
			stat.cnt++
		} else {
			stats[name] = &stationStat{
				min: temp,
				max: temp,
				sum: int64(temp),
				cnt: 1,
			}
		}
	}
	for _, key := range slices.Sorted(maps.Keys(stats)) {
		fmt.Fprintf(w, "\"%s\"/%v\n", key, stats[key])
	}
}

func parseTemp(bytes []byte) int {
	sign := 1
	if bytes[0] == '-' {
		sign = -1
		bytes = bytes[1:]
	}
	temp := 0
	for _, b := range bytes {
		if b == '.' {
			continue
		}
		temp = temp*10 + int(b-'0')
	}
	return sign*temp
}
