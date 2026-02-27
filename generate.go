package main

import (
	"bufio"
	"fmt"
	"io"
	"math/rand/v2"
	"strconv"
	"strings"
	"time"
)

type station struct {
	name     string
	meanTemp float64
}

func (s station) temp() float64 {
	return rand.NormFloat64()*7 + s.meanTemp
}

func Generate(size int, r io.Reader, w io.Writer) {
	stations, err := readStations(r)
	bw := bufio.NewWriter(w)
	defer bw.Flush()
	if err != nil {
		panic(err)
	}
	start := time.Now()
	localStart := start
	for i := range size {
		station := stations[rand.IntN(len(stations))]
		fmt.Fprintf(bw, "%s;%.2f\n", station.name, station.temp())
		if i%50_000_000 == 0 {
			fmt.Printf("Generated %d measurements in %f s\n", i, time.Since(localStart).Seconds())
			localStart = time.Now()
		}
	}
	fmt.Printf("Generated %d measurements in %f s\n", size, time.Since(start).Seconds())
}

func readStations(r io.Reader) ([]station, error) {
	stations := make([]station, 0)
	s := bufio.NewScanner(r)
	for s.Scan() {
		line := s.Text()
		if strings.HasPrefix(line, "#") {
			continue
		}
		parts := strings.Split(line, ";")
		if len(parts[0]) > 100 {
			continue
		}
		meanTemp, err := strconv.ParseFloat(parts[1], 64)
		if err != nil {
			return nil, err
		}
		station := station{
			name:     parts[0],
			meanTemp: meanTemp,
		}
		stations = append(stations, station)
	}
	return stations, nil
}
