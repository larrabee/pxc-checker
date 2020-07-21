# Build in Docker container
FROM golang:1.13 as builder

ENV CGO_ENABLED 0
WORKDIR /src
COPY . ./
RUN go mod vendor && \
    go build -o pxc-checker ./

# Create image
FROM scratch
EXPOSE 9200
ENV WEB_LISTEN ":9200"
COPY --from=builder /src/pxc-checker /pxc-checker
ENTRYPOINT ["/pxc-checker"]
