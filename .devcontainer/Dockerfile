FROM mcr.microsoft.com/devcontainers/go:dev-1.26-bookworm
ENV DEBIAN_FRONTEND=noninteractive \
    NPM_CONFIG_UPDATE_NOTIFIER=false


RUN set -eux; \
    rm -f /etc/apt/sources.list.d/yarn.list /usr/share/keyrings/yarnkey.gpg 2>/dev/null || true; \
    apt-get update; \
    apt-get install -y --no-install-recommends \
        ca-certificates curl wget gnupg gosu dnsutils iputils-ping mc vim \
        libgtk-3-dev libwebkit2gtk-4.1-dev mingw-w64 nsis sqlite3 libsqlite3-dev xvfb; \
    update-ca-certificates; \
    curl -fsSL https://deb.nodesource.com/setup_22.x -o /tmp/nodesource_setup.sh; \
    bash /tmp/nodesource_setup.sh; \
    rm -f /tmp/nodesource_setup.sh; \
    apt-get install -y --no-install-recommends nodejs; \
    install -m 0755 -d /etc/apt/keyrings; \
    curl -fsSL https://download.docker.com/linux/debian/gpg -o /etc/apt/keyrings/docker.asc; \
    chmod a+r /etc/apt/keyrings/docker.asc; \
    printf 'deb [arch=%s signed-by=/etc/apt/keyrings/docker.asc] https://download.docker.com/linux/debian bookworm stable\n' "$(dpkg --print-architecture)" > /etc/apt/sources.list.d/docker.list; \
    apt-get update; \
    apt-get install -y --no-install-recommends docker-ce-cli docker-compose-plugin; \
    wget -q https://packages.microsoft.com/config/debian/12/packages-microsoft-prod.deb -O /tmp/packages-microsoft-prod.deb; \
    dpkg -i /tmp/packages-microsoft-prod.deb; \
    rm -f /tmp/packages-microsoft-prod.deb; \
    apt-get update; \
    apt-get install -y --no-install-recommends powershell; \
    apt-get clean; \
    rm -rf /var/lib/apt/lists/* /tmp/*

RUN set -eux; \
    npm install -g npm@11.1.0 typescript@7.0.2

RUN set -eux; \
    gosu vscode go install gotest.tools/gotestsum@v1.13.0; \
    gosu vscode go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.10.1; \
    gosu vscode go install golang.org/x/tools/gopls@v0.21.0; \
    gosu vscode go install github.com/wailsapp/wails/v2/cmd/wails@v2.17.1

RUN set -eux; \
    curl -fsSL https://github.com/nektos/act/releases/download/v0.2.94/act_Linux_x86_64.tar.gz -o /tmp/act.tar.gz; \
    tar -xzf /tmp/act.tar.gz -C /usr/local/bin act; \
    chmod +x /usr/local/bin/act; \
    rm -f /tmp/act.tar.gz

RUN set -eux; \
    export HOME=/tmp/opencode-home; \
    mkdir -p "$HOME"; \
    curl -fsSL https://opencode.ai/install | bash; \
    install -m 0755 "$HOME/.opencode/bin/opencode" /usr/local/bin/opencode; \
    rm -rf "$HOME"

RUN set -eux; \
    mkdir -p /workspaces/Lyyyra/frontend; \
    chown -R vscode:vscode /workspaces/Lyyyra

RUN set -eux; mkdir -p /etc/lyyyra
COPY lyyyra-history.sh /etc/lyyyra/lyyyra-history.sh
RUN set -eux; \
    mkdir -p /home/vscode/.bash_history_data; \
    printf '%s\n' '. /etc/lyyyra/lyyyra-history.sh' >> /home/vscode/.bashrc; \
    chown vscode:vscode /home/vscode/.bashrc; \
    chown -R vscode:vscode /home/vscode/.bash_history_data

ENV DEBIAN_FRONTEND=
