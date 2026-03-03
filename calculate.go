package main

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"math"
	"os"
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

type stationStat struct {
	name []byte
	min  int
	max  int
	sum  int64
	cnt  int
	hash uint64
	psl  int
}

func process(b []byte, w io.Writer) {
	stats := make([]*stationStat, 1<<14)

	for {
		lb := bytes.IndexByte(b, '\n')
		if lb < 0 {
			break
		}
		l := b[:lb]
		b = b[lb+1:]

		name, tempStr, _ := bytes.Cut(l, []byte{';'})
		temp := parseTemp(tempStr)

		h := hash(name)
		if stat, ok := lookup(stats, name, h); ok {
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
			}
			insert(stats, h, v)
		}
	}

	bw := bufio.NewWriter(w)
	defer bw.Flush()
	for _, stat := range stats {
		if stat != nil {
			mn := float64(stat.min) / 10
			mx := float64(stat.max) / 10
			avg := float64(stat.sum) / 10 / float64(stat.cnt)
			fmt.Fprintf(bw, "\"%s\"/%.1f/%.1f/%.1f\n", stat.name, mn, mx, avg)
		}
	}
}

func hash(l []byte) uint64 {
	h := uint64(fnvOffset64)
	for _, r := range l {
		h ^= uint64(r)
		h *= fnvPrime64
	}
	return h
}

func parseTemp(tempStr []byte) int {
	sign := 1
	if tempStr[0] == '-' {
		sign = -1
		tempStr = tempStr[1:]
	}
	temp := 0
	for _, r := range tempStr {
		if r == '.' {
			continue
		}
		temp = temp*10 + int(r-'0')
	}
	temp = sign * temp
	return temp
}

func lookup(stats []*stationStat, name []byte, h uint64) (*stationStat, bool) {
	l := uint64(len(stats))
	idx := h & (l - 1)
	ok := false
	psl := 0
	for stats[idx] != nil {
		if stats[idx].hash == h && bytes.Equal(stats[idx].name, name) {
			ok = true
			break
		}
		if psl > stats[idx].psl {
			break
		}
		idx = (idx + 1) & (l - 1)
		psl++
	}
	return stats[idx], ok
}

func insert(stats []*stationStat, h uint64, v *stationStat) {
	l := uint64(len(stats))
	idx := h & (l - 1)
	vpsl := 0
	v.hash = h
	for stats[idx] != nil {
		if vpsl > stats[idx].psl {
			v.psl = vpsl
			stats[idx], v = v, stats[idx]
			vpsl = v.psl
		}
		idx = (idx + 1) & (l - 1)
		vpsl++
	}
	v.psl = vpsl
	stats[idx] = v
}

