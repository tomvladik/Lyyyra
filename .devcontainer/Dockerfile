FROM mcr.microsoft.com/devcontainers/go:dev-1.26-bookworm
ENV PATH=$PATH:/opt/node/bin
ENV DEBIAN_FRONTEND=noninteractive

# COPY corporate-root-ca.crt /usr/local/share/ca-certificates/corporate-root-ca.crt
RUN update-ca-certificates
# ENV NODE_EXTRA_CA_CERTS=/usr/local/share/ca-certificates/corporate-root-ca.crt

# Remove stale Yarn apt repo (shipped by base image) whose GPG key has expired
RUN rm -f /etc/apt/sources.list.d/yarn.list /usr/share/keyrings/yarnkey.gpg 2>/dev/null || true

RUN apt-get update

RUN set -eux; \
    curl -fsSL https://deb.nodesource.com/setup_22.x -o /tmp/nodesource_setup.sh; \
    bash /tmp/nodesource_setup.sh; \
    rm -f /tmp/nodesource_setup.sh; \
    apt-get install -y nodejs=22.23.4-1nodesource1; \
    npm --version

RUN set -eux; \
    npm install -g npm@11.1.0; \
    npm install -g typescript@7.0.2;

RUN set -eux; \
    apt-get update; \
    apt-get install -y --fix-missing dnsutils iputils-ping gosu mc vim 

RUN set -eux; \
    apt-get install -y libgtk-3-dev=3.24.38-2~deb12u3 libwebkit2gtk-4.1-dev=2.50.6-1~deb12u2

RUN set -eux; \
    apt-get install -y mingw-w64=10.0.0-3

RUN set -eux; \
    apt-get update; \
    apt-get install -y nsis=3.08-3+deb12u1 sqlite3=3.40.1-2+deb12u2 libsqlite3-dev=3.40.1-2+deb12u2

RUN set -eux; \
    apt-get install -y xvfb=2:21.1.7-3+deb12u13

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
    gosu vscode go install gotest.tools/gotestsum@v1.13.0; \
    gosu vscode go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.10.1; \
    gosu vscode go install golang.org/x/tools/gopls@v0.21.0
RUN set -eux; \
    curl -fsSL https://github.com/nektos/act/releases/download/v0.2.94/act_Linux_x86_64.tar.gz -o /tmp/act.tar.gz; \
    tar -xzf /tmp/act.tar.gz -C /usr/local/bin act; \
    chmod +x /usr/local/bin/act; \
    rm -f /tmp/act.tar.gz

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
    gosu vscode go install github.com/wailsapp/wails/v2/cmd/wails@v2.17.1; \
    gosu vscode curl -fsSL https://opencode.ai/install | bash \
    && mv $HOME/.opencode/bin/opencode /usr/local/bin/opencode

# Install frontend coverage tool during build
RUN set -eux; \
    mkdir -p /workspaces/Lyyyra/frontend; \
    chown -R vscode:vscode /workspaces/Lyyyra

ENV DEBIAN_FRONTEND=
