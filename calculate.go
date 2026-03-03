package main

import (
	"bytes"
	"io"
	"math"
	"os"
	"strconv"
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
}

func process(b []byte, w io.Writer) {
	stats := make([]*stationStat, 1<<15)

	for len(b) > 0 {
		i := 0
		var name []byte
		h := uint64(fnvOffset64)
		for ; i < len(b); i++ {
			r := b[i]
			h ^= uint64(r)
			h *= fnvPrime64
			if r == ';' {
				name = b[:i]
				break
			}
		}
		i++
		sign := 1
		if b[i] == '-' {
			sign = -1
			i++
		}
		temp := 0
		for ; i < len(b); i++ {
			r := b[i]
			if r == '.' {
				continue
			}
			if r == '\n' {
				b = b[i+1:]
				break
			}
			temp = temp*10 + int(r-'0')
		}
		temp = sign * temp

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

	bw := make([]byte, 0, 1024*1024)
	for _, stat := range stats {
		if stat != nil {
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
	}
	w.Write(bw)
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

