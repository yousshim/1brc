package main

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
)

type stationStat struct {
	name []byte
	min  int
	max  int
	sum  int64
	cnt  int
	hash uint64
}

func (stat stationStat) String() string {
	mean := float64(stat.sum) / 10 / float64(stat.cnt)
	minn := float64(stat.min) / 10
	maxx := float64(stat.max) / 10
	return fmt.Sprintf("\"%s\"/%.1f/%.1f/%.1f", string(stat.name), minn, maxx, mean)
}

func Calculate(r io.Reader, w io.Writer) {
	s := bufio.NewScanner(r)
	stats := make([]*stationStat, 1<<17)
	for s.Scan() {
		line := s.Bytes()
		splitIdx := bytes.IndexRune(line, ';')
		name := line[:splitIdx]
		temp := parseTemp(line[splitIdx+1:])
		h := hash(name)
		if idx, ok := probe(stats, name, h); ok {
			stat := stats[idx]
			stat.min = min(stat.min, temp)
			stat.max = max(stat.max, temp)
			stat.sum += int64(temp)
			stat.cnt++
		} else {
			stat := &stationStat{
				name: append([]byte{}, name...),
				min:  temp,
				max:  temp,
				sum:  int64(temp),
				cnt:  1,
			}
			insert(stats, stat, h)
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
	return sign * temp
}

func hash(bytes []byte) uint64 {
	const (
		fnvOffset64 = 14695981039346656037
		fnvPrime64  = 1099511628211
	)

	h := uint64(fnvOffset64)
	for _, b := range bytes {
		h ^= uint64(b)
		h *= fnvPrime64
	}
	return h
}

func probe(table []*stationStat, k []byte, h uint64) (uint64, bool) {
	idx := h & (uint64(len(table)) - 1)
	for i := idx; i < uint64(len(table)); i++ {
		if table[i] == nil {
			return i, false
		}
		if table[i].hash == h && bytes.Equal(table[i].name, k) {
			return i, true
		}
	}
	for i := range idx {
		if table[i] == nil {
			return i, false
		}
		if table[i].hash == h && bytes.Equal(table[i].name, k) {
			return i, true
		}
	}
	panic("unreachable")
}

func insert(table []*stationStat, stat *stationStat, h uint64) {
	stat.hash = h
	n := uint64(len(table))
	idx := h & (n - 1)
	dist := uint64(0)

	for {
		if table[idx] == nil {
			table[idx] = stat
			return
		}

		existing := table[idx]
		existingDist := (idx + n - existing.hash) & (n - 1)
		if existingDist < dist {
			table[idx], stat = stat, existing
			dist = existingDist
		}

		idx = (idx + 1) & (n - 1)
		dist++
		if dist >= n {
			panic("hash table is full")
		}
	}
}
