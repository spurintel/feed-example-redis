# feed-example-redis
This is a fully working sample program to ingest Spur feeds into a redis database.

The go binary provides 3 commands:
1. daemon - runs forever, checks for latest feed and inserts it into redis, updates using realtime data
2. insert - inserts a feed file into redis and exits
3. merge - merges a realtime file into redis and exits

## Requirements
To run this program, you will need the following:

* Go
* Docker
* A Spur token with access to feeds and realtime data. It should exposed in your environment as "SPUR_REDIS_API_TOKEN"

## Running
You have 2 options for running the application. You can build the binary directly with make and run it yourself.
Alternatively, you can build and run it in containers with docker compose.

### Option 1 - run with docker compose
```bash
# build and run with docker compose
cd feed-example-redis
make run
```

### Option 2 - build an run the binary directly
```bash
# build the binary
cd feed-example-redis
make

# Start a redis server
docker run --rm -p 6379:6379 --name redis redis:latest

# Run the binary in daemon mode, assumes you have you token in a .env file
export $(cat .env | xargs) && ./target/spurredis_darwin_arm64 daemon
```

### Querying for the data 
```bash
docker exec -it redis redis-cli GET 1.2.3.4
```
