# syntax=docker/dockerfile:1

FROM golang:1.25-alpine AS build
WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/termsnake ./cmd/termsnake

FROM alpine:3.20
RUN apk add --no-cache ncurses-terminfo-base \
    && adduser -D -h /home/termsnake termsnake
ENV LANG=C.UTF-8
WORKDIR /home/termsnake
USER termsnake

COPY --from=build /out/termsnake /usr/local/bin/termsnake
ENTRYPOINT ["termsnake"]
