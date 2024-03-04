# feed-example-redis
This is a fully working sample program designed to ingest Spur feeds into a Redis database.

The Go binary provides 3 commands:
1. **daemon** - Runs indefinitely, checks for the latest feed, and inserts it into Redis, updates using real-time data if your token supports it.
2. **insert** - Inserts a feed file into Redis and exits.
3. **merge** - Merges a real-time file into Redis and exits.

## Requirements
To run this program, you will need:

* Go installed on your machine.
* Docker for managing containers.
* A Spur token with access to feeds and optionally real-time data. This should be exposed in your environment as \`SPUR_REDIS_API_TOKEN\`.

## Quickstart
To just run the app with a redis server, you can use the following Docker Compose file. This will start a Redis server and the feed-example-redis app. 
Please see the [Configuration](#configuration) section for more information on the environment variables.
NOTE: You will need to set the `SPUR_REDIS_API_TOKEN` environment variable to your Spur token.

```bash
version: '3'

services:
  redis:
    image: redis
    ports:
      - "6379:6379"
    healthcheck:
      test: [ "CMD", "redis-cli", "--raw", "incr", "ping" ]
      interval: 30s
      timeout: 10s
      retries: 5

  spurredis:
    image: spurintelligence/spurredis:latest
    depends_on:
      redis:
        condition: service_healthy
    ports:
      - "8080:8080"
    environment:
      SPUR_REDIS_API_TOKEN: ${SPUR_REDIS_API_TOKEN}
      SPUR_REDIS_CHUNK_SIZE: 5000
      SPUR_REDIS_CONCURRENT_NUM: 4
      SPUR_REDIS_ADDR: redis:6379
      SPUR_REDIS_LOCAL_API_AUTH_TOKENS: testtoken1,testtoken2
```

Then run the following command to test the API:
```bash
curl -vv -H "TOKEN: testtoken1" localhost:8080/v2/context/${YOUR_IP} | jq
```

## Running
You have two options for running the application: building the binary directly with make or using Docker Compose for containerized deployment.

### Option 1 - Run with Docker Compose
```bash
# Build and run with Docker Compose
cd feed-example-redis
make run
```

### Option 2 - Build and run the binary directly
```bash
# Build the binary
cd feed-example-redis
make

# Start a Redis server
docker run --rm -p 6379:6379 --name redis redis:latest

# Run the binary in daemon mode, assumes you have your token and other configurations set in a .env file
export $(cat .env | xargs) && ./target/spurredis_darwin_arm64 daemon
```

## Configuring and Running the API Locally
To run the API server locally, use the \`-api\` flag when starting the binary in daemon mode. This will start the local API server along with the daemon process:

```bash
# Run the binary with the API server enabled
export \$(cat .env | xargs) && ./target/spurredis_darwin_arm64 -api daemon
```

Ensure the environment variables are set correctly, especially \`SPUR_REDIS_LOCAL_API_AUTH_TOKENS\`, to use the API authentication.

## API Usage Examples
Below are examples of how to interact with the API using curl:

### Get context for an IP address
Replace `your_auth_token` with one of your valid API tokens and `your_ip_address` with the IP address you want to query.

```bash
curl -H "TOKEN: your_auth_token" http://localhost:PORT/v2/context/your_ip_address
```

Make sure to replace \`PORT\` with the actual port number your API server is listening on.

## Configuration
The application can be configured through the following environment variables:

- `SPUR_REDIS_CHUNK_SIZE`: Sets the chunk size for Redis operations.
- `SPUR_REDIS_TTL`: Sets the TTL for Redis keys.
- `SPUR_REDIS_ADDR`: Sets the Redis server address.
- `SPUR_REDIS_PASS`: Sets the Redis password if required.
- `SPUR_REDIS_DB`: Selects the Redis database.
- `SPUR_REDIS_CONCURRENT_NUM`: Sets the number of concurrent operations.
- `SPUR_REDIS_API_TOKEN`: Sets the API token for Spur.
- `SPUR_REDIS_FEED_TYPE`: Sets the feed type for Spur: anonymous, anonymous-residential.
- `SPUR_REDIS_PORT`: Sets the port for the HTTP/HTTPS server.
- `SPUR_REDIS_CERT_FILE` and `SPUR_REDIS_KEY_FILE`: Set the paths to your SSL certificate and key files for HTTPS support.
- `SPUR_REDIS_LOCAL_API_AUTH_TOKENS`: Sets the API tokens for the local api server authentication.

### Querying for the Data
```bash
docker exec -it redis redis-cli GET 1.2.3.4
```

Ensure all sensitive information and configurations are securely stored and not exposed unnecessarily.
