
## Run unit tests
go test ./... --cover

## Testing - should find 4 pages 
python3 -m http.server -d ./static
timeout 5 ./run.sh http://localhost:8000/index.html

## Execution
- arg1 Fully qualified URL
- arg2 OPTIONAL: number of sites to crawl before stopping

docker run -it --rm crawler-app ./run.sh https://www.rescale.com 1000