# xm
a simple Company usecase.
The server listens on :8080

1. MongoDb is run in a container
   a. to launch mongodb container run docker-compose up -d 
2. run the main server (xm company)
   a. go mod tidy
   b. go build
   b. ./xm
3. test 
   IMP: to run integration test (currently supported), the mongoDB container must be running
   a. go test ./...
