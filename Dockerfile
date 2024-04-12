FROM golang:alpine

WORKDIR /app

RUN apk add --no-cache gcc musl-dev

COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN go build -o ./bin/api ./src/auth-reg \
    && go build -o ./bin/migrate ./src/migrate

CMD ["/app/bin/api"]

EXPOSE 8082
