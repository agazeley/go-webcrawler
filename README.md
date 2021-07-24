## Run unit tests
- go test ./... --cover

## Testing - should find 4 pages, requires running two processs in separate terminals
- python3 -m http.server -d ./static
- timeout 5 ./run.sh http://localhost:8000/index.html

## Execution
- arg1 Fully qualified URL
- arg2 OPTIONAL: number of sites to crawl before stopping
./run.sh <ARG1> <ARG2>