FROM golang:1.20

# Set the environment variables
ENV SPUR_REDIS_API_TOKEN=$SPUR_REDIS_API_TOKEN
ENV SPUR_REDIS_CHUNK_SIZE=$SPUR_REDIS_CHUNK_SIZE
ENV SPUR_REDIS_CONCURRENT_NUM=$SPUR_REDIS_CONCURRENT_NUM

COPY ./target/spurredis_linux_amd64 /root/spurredis

CMD ["/root/spurredis", "daemon"]