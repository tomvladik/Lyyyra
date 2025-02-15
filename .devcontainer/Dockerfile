FROM mcr.microsoft.com/devcontainers/go:1-1.22-bookworm
ENV PATH=$PATH:/opt/node/bin
ENV DEBIAN_FRONTEND=noninteractive

RUN apt-get update && apt-get upgrade -y

RUN set -eux; \
    curl -fsSL https://deb.nodesource.com/setup_20.x | sudo bash -; \
    sudo apt-get install -y nodejs

RUN set -eux; \
    npm install -g npm@11.1.0; \
    npm install -g esbuild; \
    npm install -g typescript;

RUN set -eux; \
    apt-get install -y dnsutils iputils-ping gosu mc vim libgtk-3-dev libwebkit2gtk-4.0-dev nsis && \
    rm -rf /var/lib/apt/lists/*

ENV DEBIAN_FRONTEND=

RUN set -eux; \
    gosu vscode go install github.com/wailsapp/wails/v2/cmd/wails@latest    