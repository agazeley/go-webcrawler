
## Run unit tests
chmod +x test.sh && ./test.sh

## Execution
- arg1 Fully qualified URL
- arg2 OPTIONAL: number of sites to crawl before stopping

docker run -it --rm crawler-app ./run.sh https://www.rescale.com 1000