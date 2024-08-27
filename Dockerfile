FROM alpine:latest

WORKDIR /discord-bot

COPY build/discord-bot-alpine .config.json .env ./

RUN apk update && apk upgrade && apk add --no-cache ffmpeg

EXPOSE 8080

CMD [ "./discord-bot-alpine" ]