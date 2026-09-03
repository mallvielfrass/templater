 
build:
	go build -o main.bin cmd/app/*.go 
test-dimension:
	go test -count=1 -v -run ^TestDimensionSuite$$ ./internal/exelReader

test-exdocConverter:
	go test -count=1 -v -run ^TestExDocConverter_Integration$$ ./internal/exdocConverter