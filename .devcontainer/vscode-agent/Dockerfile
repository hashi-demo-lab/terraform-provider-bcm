FROM node:22-slim

# Set all build arguments at once
ARG TZ \
    USERNAME=node \
    TERRAFORM_VERSION="1.13.5" \
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

# Install packages, configure locale, set up directories, and install GitHub Copilot in one layer
# Note: node:22-slim requires ca-certificates for HTTPS operations
RUN apt update && apt install -y --no-install-recommends \
    ca-certificates less git procps sudo fzf zsh man-db unzip gnupg2 gh \
    python3 python3-pip python3-venv python3-full \
    jq curl locales wget software-properties-common && \
    # Configure locale
    sed -i 's/^# *en_US.UTF-8 UTF-8/en_US.UTF-8 UTF-8/' /etc/locale.gen && \
    locale-gen en_US.UTF-8 && \
    update-locale LANG=en_US.UTF-8 LC_ALL=en_US.UTF-8 && \
    # Set up directories and permissions
    mkdir -p /usr/local/share/npm-global /workspace && \
    chown -R node:node /usr/local/share /workspace && \
    # Install GitHub Copilot CLI
    npm install -g @github/copilot && \
    npm cache clean --force && \
    # Clean up apt cache
    rm -rf /var/lib/apt/lists/*

# Set all environment variables at once
ENV TZ="$TZ" \
    LANG=en_US.UTF-8 \
    LC_ALL=en_US.UTF-8 \
    LANGUAGE=en_US:en \
    NPM_CONFIG_PREFIX=/usr/local/share/npm-global \
    PATH=/usr/local/share/npm-global/bin:/home/node/.local/bin:$PATH \
    SHELL=/bin/bash

WORKDIR /workspace

# Configure sudo and install uv system-wide in one layer as root
RUN groupadd -f wheel && \
    usermod -aG wheel node && \
    echo "%wheel ALL=(ALL) NOPASSWD: ALL" >> /etc/sudoers.d/wheel-group && \
    chmod 0440 /etc/sudoers.d/wheel-group && \
    # Install uv system-wide
    curl -LsSf https://astral.sh/uv/install.sh | sh

# Copy library scripts and set permissions as root
COPY library-scripts/*.sh /tmp/library-scripts/
RUN chmod +x /tmp/library-scripts/*.sh

# Switch to node user for remaining installations
USER node

# Install uv for user, Python tools, and run library scripts in one layer
RUN curl -LsSf https://astral.sh/uv/install.sh | sh && \
    echo 'export PATH=$HOME/.local/bin:$PATH' >> ~/.bashrc && \
    # Install Python tools
    $HOME/.local/bin/uv tool install pre-commit && \
    $HOME/.local/bin/uv tool install specify-cli --from git+https://github.com/github/spec-kit.git && \
    # Install cloud CLI tools and terraform tools
    #/tmp/library-scripts/cloud-cli-tools.sh && \
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
    # Clean up
    sudo rm -rf /tmp/library-scripts