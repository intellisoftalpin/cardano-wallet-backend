# Build
FROM golang:1.19-alpine as builder
WORKDIR /build
COPY go.* ./
RUN go mod download
COPY . ./
RUN go build -o /cardano-wallet-backend

# Copy
FROM alpine:3
COPY --from=builder cardano-wallet-backend /bin/cardano-wallet-backend
ENTRYPOINT ["/bin/cardano-wallet-backend"]
