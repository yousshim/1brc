package main

import (
	"bytes"
	"io"
	"math"
	"os"
	"runtime"
	"slices"
	"strconv"
	"sync"
	"syscall"
)

func Calculate(f *os.File, w io.Writer) {
	fi, err := f.Stat()
	if err != nil {
		println(err.Error())
		os.Exit(1)
	}
	s := fi.Size()
	if s == 0 {
		println("empty file")
		os.Exit(0)
	}
	if s > math.MaxInt {
		println("file too large")
		os.Exit(1)
	}

	d, err := syscall.Mmap(int(f.Fd()), 0, int(s), syscall.PROT_READ, syscall.MAP_SHARED)
	if err != nil {
		println(err.Error())
		os.Exit(1)
	}
	defer syscall.Munmap(d)

	process(d, w)
}

const (
	fnvOffset64 = 14695981039346656037
	fnvPrime64  = 1099511628211
)

var NEW_LINE = []byte{'\n'}
var SIMI_COLON = []byte{';'}

type stationStat struct {
	name []byte
	min  int
	max  int
	sum  int64
	cnt  int
	hash uint64
}

func process(b []byte, w io.Writer) {
	workers := runtime.NumCPU()
	if workers < 2 || len(b) < workers {
		workers = 1
	}

	chunks := splitChunks(b, workers)
	chunkStats := make([][]*stationStat, len(chunks))
	wg := sync.WaitGroup{}
	wg.Add(len(chunks))

	for i, chunk := range chunks {
		go func() {
			defer wg.Done()
			chunkStats[i] = processChunk(chunk)
		}()
	}
	wg.Wait()

	stats := make([]*stationStat, 1<<15)
	for _, chunkStat := range chunkStats {
		mergeStats(stats, chunkStat)
	}

	slices.SortFunc(stats, func(a, b *stationStat) int {
		if a == nil {
			return 1
		}
		if b == nil {
			return -1
		}
		return bytes.Compare(a.name, b.name)
	})

	bw := make([]byte, 0, 1024*1024)
	for _, stat := range stats {
		if stat == nil {
			break
		}
		mn := float64(stat.min) / 10
		mx := float64(stat.max) / 10
		avg := float64(stat.sum) / 10 / float64(stat.cnt)
		bw = append(bw, '"')
		bw = append(bw, stat.name...)
		bw = append(bw, '"')
		bw = append(bw, '/')
		bw = strconv.AppendFloat(bw, mn, 'f', 1, 64)
		bw = append(bw, '/')
		bw = strconv.AppendFloat(bw, mx, 'f', 1, 64)
		bw = append(bw, '/')
		bw = strconv.AppendFloat(bw, avg, 'f', 1, 64)
		bw = append(bw, '\n')
	}
	w.Write(bw)
}

func splitChunks(b []byte, workers int) [][]byte {
	if workers <= 1 {
		return [][]byte{b}
	}

	target := len(b)
	if workers > target {
		workers = target
	}

	chunks := make([][]byte, 0, workers)
	avg := target / workers
	start := 0
	for i := 0; i < workers && start < len(b); i++ {
		end := start + avg
		if i == workers-1 || end >= len(b) {
			end = len(b)
		} else {
			for end < len(b) && b[end-1] != '\n' {
				end++
			}
		}
		if end == start {
			continue
		}
		chunks = append(chunks, b[start:end])
		start = end
	}

	if start < len(b) {
		chunks = append(chunks, b[start:])
	}

	if len(chunks) == 0 {
		return [][]byte{b}
	}

	return chunks
}

func processChunk(b []byte) []*stationStat {
	stats := make([]*stationStat, 1<<15)
	for len(b) > 0 {
		var name, tempStr []byte
		name, b, _ = bytes.Cut(b, SIMI_COLON)
		tempStr, b, _ = bytes.Cut(b, NEW_LINE)

		temp := parseTemp(tempStr)
		h := hash(name)
		if idx, ok := lookup(stats, name, h); ok {
			stat := stats[idx]
			stat.min = min(stat.min, temp)
			stat.max = max(stat.max, temp)
			stat.sum += int64(temp)
			stat.cnt++
		} else {
			v := &stationStat{
				name: append([]byte{}, name...),
				min:  temp,
				max:  temp,
				sum:  int64(temp),
				cnt:  1,
				hash: h,
			}
			stats[idx] = v
		}
	}
	return stats
}

func parseTemp(b []byte) int {
	i := 0
	sign := 1
	if b[i] == '-' {
		sign = -1
		i++
	}
	temp := 0
	if b[i+1] == '.' {
		temp = int(b[i]) - '0'
		temp = temp*10 + int(b[i+2]) - '0'
	} else {
		temp = int(b[i]) - '0'
		temp = temp*10 + int(b[i+1]) - '0'
		temp = temp*10 + int(b[i+3]) - '0'
	}

	return sign * temp
}

func hash(name []byte) uint64 {
	h := uint64(fnvOffset64)
	for _, r := range name {
		h ^= uint64(r)
		h *= fnvPrime64
	}
	return h
}

func mergeStats(dst []*stationStat, src []*stationStat) {
	for _, stat := range src {
		if stat == nil {
			continue
		}
		if idx, ok := lookup(dst, stat.name, stat.hash); ok {
			d := dst[idx]
			d.min = min(d.min, stat.min)
			d.max = max(d.max, stat.max)
			d.sum += stat.sum
			d.cnt += stat.cnt
		} else {
			dst[idx] = &stationStat{
				name: append([]byte{}, stat.name...),
				min:  stat.min,
				max:  stat.max,
				sum:  stat.sum,
				cnt:  stat.cnt,
				hash: stat.hash,
			}
		}
	}
}

func lookup(stats []*stationStat, name []byte, h uint64) (uint64, bool) {
	l := uint64(len(stats))
	idx := h & (l - 1)
	ok := false
	for stats[idx] != nil {
		if stats[idx].hash == h && bytes.Equal(stats[idx].name, name) {
			ok = true
			break
		}
		idx = (idx + 1) & (l - 1)
	}
	return idx, ok
}
