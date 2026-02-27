.PHONY: all build gen-s calc-s gen-full calc-full clean

BIN := 1brc
STATIONS := stations.txt
SAMPLE := sample.txt
FULL := measurements.txt

all: build

build:
	go build -o $(BIN)

gen-sample: build
	./$(BIN) generate -i $(STATIONS) -o $(SAMPLE) -n 100

calc-sample: build gen-sample
	./$(BIN) calculate -i $(SAMPLE)

gen: build
	./$(BIN) generate -i $(STATIONS) -o $(FULL) -n 1000000000

calc: build gen
	./$(BIN) calculate -i $(FULL)

clean:
	rm -f $(BIN) $(SAMPLE) $(FULL)
