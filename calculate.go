package main

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
)

type stationStat struct {
	name []byte
	min int
	max int
	sum int64
	cnt int
	hash uint64
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
		splitIdx := bytes.IndexRune(line, ';')
		name := line[:splitIdx]
		idx, ok := findTableIndex(stats, name)
		temp := parseTemp(line[splitIdx+1:])
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
	bw := bufio.NewWriter(w)
	defer bw.Flush()
	for _, stat := range stats {
		if stat != nil {
			fmt.Fprintf(bw, "%v\n", stat)
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

func hash(bytes []byte, capacity int) uint64 {
	h := uint64(0)
	for _, b := range bytes {
		h = h*31 + uint64(b)
	}
	return h & (uint64(capacity) - 1)
}

func findTableIndex(table []*stationStat, key []byte) (uint64, bool) {
	h := hash(key, len(table))
	for i := h; i < uint64(len(table)); i++ {
		if table[i] == nil {
			return i, false
		}
		if table[i].hash == h {
			return i, true
		}
		if  bytes.Equal(table[i].name, key) {
			table[i].hash = h
			return i, true
		}
	}
	for i := range h {
		if table[i] == nil {
			return i, false
		}
		if table[i].hash == h {
			return i, true
		}
		if bytes.Equal(table[i].name, key) {
			table[i].hash = h
			return i, true
		}
	}
	panic("unreachable")
}

