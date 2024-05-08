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

### System Requirements

* Anonymous feeds require 5GB of memory.
* Anonymous residential feeds require a minimum 18GB of memory.

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
      SPUR_REDIS_FEED_TYPE: anonymous
      SPUR_REDIS_CHUNK_SIZE: 5000
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
# Clone the repository
git clone https://github.com/spurintel/feed-example-redis.git

# Build the binary
cd feed-example-redis
make bin

# Set the environment variables, minimum required is the Spur token
export SPUR_REDIS_API_TOKEN=your_spur_token

# Start the application with Docker Compose
docker-compose up
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

- `SPUR_REDIS_CHUNK_SIZE`: Sets the chunk size for Redis operations. (default: 5000)
- `SPUR_REDIS_TTL`: Sets the TTL (in hours) for Redis keys. (default: 24)
- `SPUR_REDIS_ADDR`: Sets the Redis server address. (default: "localhost:6379")
- `SPUR_REDIS_PASS`: Sets the Redis password. (default: "")
- `SPUR_REDIS_DB`: Sets the Redis DB. (default: 0)
- `SPUR_REDIS_CONCURRENT_NUM`: Sets the number of concurrent processes. (default: number of CPUs)
- `SPUR_REDIS_API_TOKEN`: Sets the Spur API token. (Required)
- `SPUR_REDIS_FEED_TYPE`: Sets the Spur feed type. (default: "anonymous")
- `SPUR_REDIS_REALTIME_ENABLED`: Sets whether realtime feed is enabled. (default: false)
- `SPUR_REDIS_PORT`: Sets the port for the application. (default: 8080)
- `SPUR_REDIS_CERT_FILE`: Specifies the TLS Cert file. (default: "")
- `SPUR_REDIS_KEY_FILE`: Specifies the TLS Key file. (default: "")
- `SPUR_REDIS_LOCAL_API_AUTH_TOKENS`: Sets the local API Auth tokens. (Required; Tokens are comma separated)
- `SPUR_REDIS_IPV6_NETWORK_FEED_BETA`: Also include data from IPv6 network info feeds (BETA). May increase resource requirements.

Please note: For SPUR_REDIS_API_TOKEN and SPUR_REDIS_LOCAL_API_AUTH_TOKENS, if these are not set, the application will not run.

### Querying for the Data
```bash
docker exec -it redis redis-cli GET 1.2.3.4
```

Ensure all sensitive information and configurations are securely stored and not exposed unnecessarily.
