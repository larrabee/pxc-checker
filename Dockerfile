# Build in Docker container
FROM golang:1.13 as builder

ENV CGO_ENABLED 0
WORKDIR /src
COPY . ./
RUN go mod vendor && \
    go build -o pxc-checker ./

# Create s3sync image
FROM scratch
COPY --from=builder /src/pxc-checker /pxc-checker
ENTRYPOINT ["/pxc-checker"]
