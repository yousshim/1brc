package main

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"slices"
)

type stationStat struct {
	name []byte
	min int
	max int
	sum int64
	cnt int
}

func (stat stationStat) String() string {
	mean := float64(stat.sum)/10/float64(stat.cnt)
	minn := float64(stat.min)/10
	maxx := float64(stat.max)/10
	return fmt.Sprintf("\"%s\"/%.1f/%.1f/%.1f", string(stat.name), minn, maxx, mean)
}

func Calculate(r io.Reader, w io.Writer) {
	s := bufio.NewScanner(r)
	stats := make([]*stationStat, 1<<17)
	for s.Scan() {
		line := s.Bytes()
		splitIdx := slices.Index(line, ';')
		name := line[:splitIdx]
		temp := parseTemp(line[splitIdx+1:])
		idx, ok := find(stats, name)
		if ok {
			stat := stats[idx]
			stat.min = min(stat.min, temp)
			stat.max = max(stat.max, temp)
			stat.sum += int64(temp)
			stat.cnt++
		} else {
			stats[idx] = &stationStat{
				name: append([]byte{}, name...),
				min: temp,
				max: temp,
				sum: int64(temp),
				cnt: 1,
			}
		}
	}
	for _, stat := range stats {
		if stat != nil {
			fmt.Fprintf(w, "%v\n", stat)
		}
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

func hash(key []byte) int {
	h := 0
	for _, b := range key {
		h = h*31 + int(b)
	}
	if h < 0 {
		h = -h
	}
	return h
}

func find(table []*stationStat, key []byte) (int, bool) {
	h := hash(key) % len(table)
	if table[h] == nil {
		return h, false
	}
	if bytes.Equal(table[h].name, key) {
		return h, true
	}
	for i := h; i < len(table); i++ {
		if table[i] == nil {
			return i, false
		}
		if bytes.Equal(table[i].name, key) {
			return i, true
		}
	}
	for i := range h {
		if table[i] == nil {
			return i, false
		}
		if bytes.Equal(table[i].name, key) {
			return i, true
		}
	}
	panic("unreachable")
}
