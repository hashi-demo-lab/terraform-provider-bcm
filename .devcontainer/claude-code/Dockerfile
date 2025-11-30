# Copyright (c) HashiCorp, Inc.
# SPDX-License-Identifier: MPL-2.0

FROM node:22-slim

# Set all build arguments at once
ARG TZ \
    USERNAME=node \
    GO_VERSION="1.24.0" \
    TERRAFORM_VERSION="1.14.0" \
    TERRAFORM_DOCS_VERSION="0.20.0" \
    TFSEC_VERSION="1.28.14" \
    TERRASCAN_VERSION="1.19.9" \
    TFLINT_VERSION="0.59.1" \
    TFLINT_AWS_RULESET_VERSION="0.23.1" \
    TFLINT_AZURE_RULESET_VERSION="0.23.0" \
    TFLINT_GCP_RULESET_VERSION="0.23.1" \
    INFRACOST_VERSION="0.10.42" \
    CHECKOV_VERSION="3.2.489" \
    VAULT_RADAR_VERSION="0.36.0" \
    INSTALL_VAULT_RADAR="true"

# Install packages, configure locale, set up directories, and install delta in one layer
# Note: node:20-slim requires ca-certificates for HTTPS operations
RUN apt update && apt install -y --no-install-recommends \
    ca-certificates less git procps sudo zsh man-db unzip gnupg2 gh \
    iptables ipset iproute2 dnsutils aggregate \
    python3 python3-pip python3-venv python3-full \
    jq curl locales wget && \
    # Configure locale
    sed -i 's/^# *en_US.UTF-8 UTF-8/en_US.UTF-8 UTF-8/' /etc/locale.gen && \
    locale-gen en_US.UTF-8 && \
    update-locale LANG=en_US.UTF-8 LC_ALL=en_US.UTF-8 && \
    # Set up directories and permissions
    mkdir -p /usr/local/share/npm-global /commandhistory /workspace /home/node/.claude /home/node/.config && \
    touch /commandhistory/.bash_history && \
    chown -R node:node /usr/local/share /commandhistory /workspace /home/node/.claude /home/node/.config && \
    # Install delta
    ARCH=$(dpkg --print-architecture) && \
    wget -q "https://github.com/dandavison/delta/releases/download/0.18.2/git-delta_0.18.2_${ARCH}.deb" && \
    dpkg -i "git-delta_0.18.2_${ARCH}.deb" && \
    rm "git-delta_0.18.2_${ARCH}.deb" && \
    # Install Go
    GO_ARCH=$(case ${ARCH} in amd64) echo "amd64";; arm64) echo "arm64";; *) echo "amd64";; esac) && \
    wget -q "https://go.dev/dl/go${GO_VERSION}.linux-${GO_ARCH}.tar.gz" && \
    tar -C /usr/local -xzf "go${GO_VERSION}.linux-${GO_ARCH}.tar.gz" && \
    rm "go${GO_VERSION}.linux-${GO_ARCH}.tar.gz" && \
    # Clean up apt cache
    rm -rf /var/lib/apt/lists/*

# Set all environment variables at once
ENV TZ="$TZ" \
    LANG=en_US.UTF-8 \
    LC_ALL=en_US.UTF-8 \
    LANGUAGE=en_US:en \
    DEVCONTAINER=true \
    NPM_CONFIG_PREFIX=/usr/local/share/npm-global \
    GOROOT=/usr/local/go \
    GOPATH=/go \
    PATH=/usr/local/go/bin:/go/bin:/usr/local/share/npm-global/bin:/root/.local/bin:$PATH \
    SHELL=/bin/zsh

WORKDIR /workspace

# Configure sudo and install uv system-wide in one layer as root
RUN groupadd -f wheel && \
    usermod -aG wheel node && \
    echo "%wheel ALL=(ALL) NOPASSWD: ALL" >> /etc/sudoers.d/wheel-group && \
    chmod 0440 /etc/sudoers.d/wheel-group && \
    # Install uv system-wide
    curl -LsSf https://astral.sh/uv/install.sh | sh

# Copy and run library scripts
COPY library-scripts/*.sh /tmp/library-scripts/
RUN chmod +x /tmp/library-scripts/*.sh && \
    /tmp/library-scripts/terraform-tools.sh \
        "${TERRAFORM_VERSION}" \
        "${TERRAFORM_DOCS_VERSION}" \
        "${TFSEC_VERSION}" \
        "${TERRASCAN_VERSION}" \
        "${TFLINT_VERSION}" \
        "${TFLINT_AWS_RULESET_VERSION}" \
        "${TFLINT_AZURE_RULESET_VERSION}" \
        "${TFLINT_GCP_RULESET_VERSION}" \
        "${INFRACOST_VERSION}" \
        "${CHECKOV_VERSION}" \
        "${VAULT_RADAR_VERSION}" \
        "${INSTALL_VAULT_RADAR}" && \
    rm -rf /tmp/library-scripts

# Switch to node user for remaining installations
USER node

# Install zsh theme, Claude, uv for user, and Python tools in one layer
RUN sh -c "$(wget -O- https://github.com/deluan/zsh-in-docker/releases/download/v1.2.0/zsh-in-docker.sh)" -- \
        -p git \
        -a "export PROMPT_COMMAND='history -a' && export HISTFILE=/commandhistory/.bash_history" \
        -x && \
    npm install -g @anthropic-ai/claude-code ccstatusline @langfuse/tracing && \
    curl -LsSf https://astral.sh/uv/install.sh | sh && \
    echo 'export PATH="$HOME/.local/bin:$PATH"' >> ~/.zshrc && \
    ~/.local/bin/uv tool install pre-commit --with pre-commit-uv && \
    ~/.local/bin/uv tool install specify-cli --from git+https://github.com/github/spec-kit.git

# Configure ccstatusline with Powerline enabled
RUN mkdir -p /home/node/.config/ccstatusline && \
    echo '{\
  "version": 3,\
  "lines": [\
    [\
      {\
        "id": "1",\
        "type": "model",\
        "color": "cyan"\
      },\
      {\
        "id": "2",\
        "type": "separator"\
      },\
      {\
        "id": "3",\
        "type": "context-length",\
        "color": "brightBlack"\
      },\
      {\
        "id": "4",\
        "type": "separator"\
      },\
      {\
        "id": "5",\
        "type": "git-branch",\
        "color": "magenta"\
      },\
      {\
        "id": "6",\
        "type": "separator"\
      },\
      {\
        "id": "7",\
        "type": "git-changes",\
        "color": "yellow"\
      }\
    ],\
    [],\
    []\
  ],\
  "flexMode": "full-minus-40",\
  "compactThreshold": 60,\
  "colorLevel": 2,\
  "inheritSeparatorColors": false,\
  "globalBold": false,\
  "powerline": {\
    "enabled": true,\
    "separators": [\
      ""\
    ],\
    "separatorInvertBackground": [\
      false\
    ],\
    "startCaps": [""],\
    "endCaps": [""],\
    "autoAlign": true,\
    "theme": "gruvbox"\
  },\
  "defaultPadding": "     "\
}' > /home/node/.config/ccstatusline/settings.json

# Add clauded alias for claude --dangerously-skip-permissions
RUN echo 'alias clauded="claude --dangerously-skip-permissions"' >> /home/node/.zshrc && \
    chown node:node /home/node/.zshrc

# Install Claude Code hooks dependencies
COPY --chown=node:node .claude/hooks /workspace/.claude/hooks
RUN cd /workspace/.claude/hooks && npm install && npm run build
