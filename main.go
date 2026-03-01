package main

import (
	"flag"
	"fmt"
	"os"
	"runtime/pprof"
	"runtime"
)

func main() {
	var size int
	var inputFile string
	var outputFile string
	generateCmd := flag.NewFlagSet("generate", flag.ExitOnError)
	generateCmd.IntVar(&size, "n", 10, "Number of measurements to generate")
	generateCmd.StringVar(&inputFile, "i", "", "Input file with weather stations")
	generateCmd.StringVar(&outputFile, "o", "", "Output file with measurements")

	calculateCmd := flag.NewFlagSet("calculate", flag.ExitOnError)
	calculateCmd.StringVar(&inputFile, "i", "", "Input file with measurements")
	cpuProf := calculateCmd.Bool("cpuprof", false, "Enable cpu profiling")
	memProf := calculateCmd.Bool("memprof", false, "Enable memory profiling")

	switch os.Args[1] {
	case "generate":
		err := generateCmd.Parse(os.Args[2:])
		if err != nil {
			generateCmd.Usage()
			os.Exit(1)
		}
		if inputFile == "" || outputFile == "" {
			generateCmd.Usage()
			os.Exit(1)
		}
		stationsFile, err := os.Open(inputFile)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		defer stationsFile.Close()
		measurementsFile, err := os.Create(outputFile)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		defer measurementsFile.Close()
		Generate(size, stationsFile, measurementsFile)
	case "calculate":
		err := calculateCmd.Parse(os.Args[2:])
		if err != nil {
			calculateCmd.Usage()
			os.Exit(1)
		}
		if inputFile == "" {
			calculateCmd.Usage()
			os.Exit(1)
		}
		measurementsFile, err := os.Open(inputFile)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		defer measurementsFile.Close()

		if *cpuProf {
			f, _ := os.Create("cpu.prof")
			defer f.Close()
			pprof.StartCPUProfile(f)
			defer pprof.StopCPUProfile()
		}

		Calculate(measurementsFile, os.Stdout)

		if *memProf {
			mf, _ := os.Create("mem.prof")
			defer mf.Close()
			runtime.GC()
			pprof.Lookup("allocs").WriteTo(mf, 0)
		}
	default:
		fmt.Println("Unknown command")
		os.Exit(1)
	}

}
