FROM golang:1.27.3-alpine AS build

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 go build -o /pricefeed ./cmd/pricefeed

FROM alpine:3.19

RUN apk add --no-cache ca-certificates

WORKDIR /app
COPY --from=build /pricefeed /app/pricefeed

CMD ["/app/pricefeed", "-config", "/app/config.json"]
