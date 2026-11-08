# タスク用の隔離コンテナ（ADR-035 決定1）。taskctl build でビルドする。
# リポジトリの mise.toml は named context "project" から受け取る。
FROM ubuntu:24.04

ARG CLAUDE_CODE_VERSION=2.1.285

# bubblewrap・socat は Claude Code 自身のサンドボックス用（コンテナ境界に重ねる多層防御）
RUN apt-get update && apt-get install -y --no-install-recommends \
    build-essential \
    bubblewrap \
    ca-certificates \
    curl \
    dnsmasq \
    ipset \
    iptables \
    procps \
    socat \
    util-linux \
    && rm -rf /var/lib/apt/lists/*

RUN curl -fsSL https://mise.run | MISE_INSTALL_PATH=/usr/local/bin/mise sh

# ubuntu イメージ既定の uid 1000 ユーザーを dev に置き換える
RUN userdel -r ubuntu && useradd -m -u 1000 -s /bin/bash dev

ENV PATH=/home/dev/.local/bin:/home/dev/.local/share/mise/shims:$PATH \
    MISE_TRUSTED_CONFIG_PATHS=/workspace \
    DISABLE_AUTOUPDATER=1

USER dev
WORKDIR /home/dev

RUN curl -fsSL https://claude.ai/install.sh | bash -s "$CLAUDE_CODE_VERSION"

# 実行時は許可ドメイン以外に出られないため、ツールチェーンはビルド時に入れておく
COPY --from=project --chown=dev mise.toml .env.example /tmp/project/
RUN cd /tmp/project && mise trust && mise install && rm -rf /tmp/project

USER root
COPY --chmod=0755 entrypoint.sh firewall.sh /usr/local/lib/taskctl/
WORKDIR /workspace
ENTRYPOINT ["/usr/local/lib/taskctl/entrypoint.sh"]
