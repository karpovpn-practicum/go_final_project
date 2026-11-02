# Сборка: статический бинарник для Linux без cgo (драйвер modernc.org/sqlite написан на Go).
FROM golang:1.24 AS build

WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download

COPY main.go ./
COPY pkg ./pkg
RUN CGO_ENABLED=0 GOOS=linux go build -o /out/scheduler .

# Запуск: в образ попадают только бинарник и фронтенд.
FROM ubuntu:24.04

WORKDIR /app
COPY --from=build /out/scheduler ./scheduler
COPY web ./web

# Файл базы лежит в /data — туда монтируется директория с хоста.
# TODO_PASSWORD в образ не зашиваем, он передаётся при запуске через -e.
ENV TODO_PORT=7540 \
    TODO_DBFILE=/data/scheduler.db

VOLUME /data
EXPOSE 7540

CMD ["./scheduler"]
