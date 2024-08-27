#!/bin/bash

isLinux=true
output_file="build/discord-bot"

if [[ "$OSTYPE" == "msys" ]] || [[ "$OSTYPE" == "cygwin" ]] || [[ "$OSTYPE" == "win32" ]]; then
    echo "### Building for Windows"
    
    isLinux=false
    output_file="build/discord-bot.exe"
else
    echo "### Building for Linux"
fi

go build -a -gcflags=all="-l -B" -ldflags="-w -s" -o "$output_file"
upx --best --ultra-brute "$output_file"
chmod +x "$output_file"

if $isLinux; then
    echo "### Building for Linux Alpine"
    
    output_file="build/discord-bot-alpine"
    export CC=musl-gcc
    export CXX=musl-g++
    go build -a -gcflags=all="-l -B" -ldflags="-w -s" -o "$output_file"
    upx --best --ultra-brute "$output_file"
    chmod +x "$output_file"
fi

