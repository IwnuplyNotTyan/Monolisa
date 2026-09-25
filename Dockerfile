FROM golang:tip-alpine3.23 as builder

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN go build -tags ssh --ldflags="-s -w" -o ./bin/monolisa ./cmd/monolisa/main.go

FROM alpine

WORKDIR /app

COPY --from=builder /app/bin/monolisa ./monolisa

EXPOSE 23234

CMD ["monolisa"]