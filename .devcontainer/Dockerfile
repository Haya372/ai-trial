FROM mcr.microsoft.com/devcontainers/base:ubuntu-24.04

RUN apt-get update && apt-get install -y \
    build-essential \
    bubblewrap \
    socat \
    && rm -rf /var/lib/apt/lists/*

RUN curl https://mise.run | MISE_INSTALL_PATH=/usr/local/bin/mise sh \
    && echo 'eval "$(mise activate bash)"' >> /etc/bash.bashrc \
    && echo 'eval "$(mise activate zsh)"' >> /etc/zsh/zshrc
