FROM golang:1.22-alpine AS buildimage

WORKDIR /discord-bot

COPY go.mod go.sum ./
RUN go mod download && go mod verify

COPY . .
RUN go build -a -gcflags=all="-l -B" -ldflags="-w -s" -o "../discord-bot"

FROM alpine:3.13

LABEL version="1.0"
LABEL description="Discord bot build with golang"
LABEL org.opencontainers.image.authors="alabsi91"

RUN apk update && apk add --no-cache ffmpeg shadow && \
    echo "**** create abc user and make our folders ****" && \
    groupmod -g 1000 users && \
    useradd -u 911 -U -d /config -s /bin/false abc && \
    usermod -G users abc && \
    mkdir -p /config && \
    echo "**** cleanup ****" && \
    rm -rf /tmp/*

COPY --from=buildimage /discord-bot/discord-bot /discord-bot/discord-bot

ENTRYPOINT [ "/discord-bot/discord-bot" ]