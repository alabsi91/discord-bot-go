FROM alpine:3.13

LABEL version="1.0"
LABEL description="Discord bot build with golang"
LABEL org.opencontainers.image.authors="alabsi91"

WORKDIR /discord-bot

RUN apk update && apk upgrade && apk add --no-cache wget tar git ffmpeg shadow && \
    # install golang
    wget https://go.dev/dl/go1.23.1.linux-amd64.tar.gz && \ 
    tar -C /usr/local -xzf go1.23.1.linux-amd64.tar.gz && \
    export PATH=$PATH:/usr/local/go/bin && \
    rm go1.23.1.linux-amd64.tar.gz && \
    # clone and build discord-bot
    git clone https://github.com/alabsi91/discord-bot-go.git && cd discord-bot-go && \
    go build -a -gcflags=all="-l -B" -ldflags="-w -s" -o "../discord-bot" && \
    # clean up
    cd .. && \
    rm -rf discord-bot-go && \
    apk del wget tar git && \
    rm -rf /usr/local/go && \
    rm -rf ~/go && \
    rm -rf /root/.cache && \
    rm -rf /var/cache/apk/*

CMD ["sh", "-c", "groupadd -g $GID appgroup && useradd -r -u $UID -g appgroup appuser && chown -R appuser:appgroup /discord-bot; exec su appuser -c './discord-bot'"]