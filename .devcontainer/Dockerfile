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
    apt-get update; \
    apt-get install -y nsis sqlite3 libsqlite3-dev

RUN set -eux; \
    apt-get install -y xvfb

RUN set -eux; \
    apt-get update; \
    apt-get install -y ca-certificates curl; \
    install -m 0755 -d /etc/apt/keyrings; \
    curl -fsSL https://download.docker.com/linux/debian/gpg -o /etc/apt/keyrings/docker.asc; \
    chmod a+r /etc/apt/keyrings/docker.asc; \
    echo "deb [arch=$(dpkg --print-architecture) signed-by=/etc/apt/keyrings/docker.asc] https://download.docker.com/linux/debian bookworm stable" | tee /etc/apt/sources.list.d/docker.list > /dev/null; \
    apt-get update; \
    apt-get install -y docker-ce-cli docker-compose-plugin

RUN set -eux; \
    usermod -aG docker vscode || true

RUN set -eux; \
    gosu vscode go install gotest.tools/gotestsum@v1.11.0; \
    gosu vscode go install github.com/golangci/golangci-lint/cmd/golangci-lint@v1.54.2; \
    gosu vscode go install golang.org/x/tools/gopls@v0.16.2

RUN set -eux; \
    curl -fsSL https://raw.githubusercontent.com/nektos/act/master/install.sh | bash

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

# Install frontend coverage tool during build
RUN set -eux; \
    mkdir -p /workspaces/Lyyyra/frontend; \
    chown -R vscode:vscode /workspaces/Lyyyra

ENV DEBIAN_FRONTEND=
