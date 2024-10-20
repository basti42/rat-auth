# build stage image
FROM golang:1.23 AS builder

# setting working dir inside the image
WORKDIR /app

# copy files
COPY go.mod go.sum ./
RUN go mod download

COPY . .

# build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a --installsuffix cgo -o main .

# prod stage image
FROM alpine:3.20 AS prod

# install necessary certs for https
RUN apk --no-cache add ca-certificates

WORKDIR /root/

# copy compiled binary from builder
COPY --from=builder /app/main .

# command to run the application
CMD ["./main"]
