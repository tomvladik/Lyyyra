FROM mcr.microsoft.com/devcontainers/go:1-1.22-bookworm
ENV PATH=$PATH:/opt/node/bin
ENV DEBIAN_FRONTEND=noninteractive

RUN apt-get update && apt-get upgrade -y

RUN set -eux; \
    curl -fsSL https://deb.nodesource.com/setup_20.x | sudo bash -; \
    sudo apt-get install -y nodejs

RUN set -eux; \
    npm install -g npm@11.1.0; \
    npm install -g typescript;

RUN set -eux; \
    apt-get update; \
    apt-get install -y --fix-missing dnsutils iputils-ping gosu mc vim 

RUN set -eux; \
    apt-get install -y libgtk-3-dev libwebkit2gtk-4.1-dev

RUN set -eux; \
    apt-get install -y mingw-w64

RUN set -eux; \
    apt-get install -y nsis sqlite3 libsqlite3-dev

RUN set -eux; \
    apt-get install -y xvfb

RUN set -eux; \
    apt-get update; \
    apt-get install -y wget; \
    wget -q https://packages.microsoft.com/config/debian/12/packages-microsoft-prod.deb; \
    dpkg -i packages-microsoft-prod.deb; \
    apt-get update; \
    apt-get install -y powershell; \
    rm packages-microsoft-prod.deb

RUN set -eux; \
    rm -rf /var/lib/apt/lists/*

RUN set -eux; \
    gosu vscode go install github.com/wailsapp/wails/v2/cmd/wails@latest

ENV DEBIAN_FRONTEND=
