ifeq ($(OS),Windows_NT)
EXT =.exe
else
EXT =
endif

all : aozora-collector$(EXT) aozora-search$(EXT)

aozora-collector$(EXT) : ./cmd/aozora-collector/main.go
	go build -o aozora-collector$(EXT) ./cmd/aozora-collector

aozora-search$(EXT) : ./cmd/aozora-search/main.go
	go build -o aozora-search$(EXT) ./cmd/aozora-search

.PHONY: clean

clean:
	rm aozora-collector$(EXT) aozora-search$(EXT)
