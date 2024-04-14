FROM golang:alpine

RUN apk add --no-cache gcc musl-dev

WORKDIR /src

COPY ./go.mod ./go.sum ./

RUN go mod download

COPY ./ ./

RUN go build -o ./bin/api ./src/avito \
    && go build -o ./bin/migrate ./src/migrate

CMD ["/src/bin/api"]
