#!/bin/bash
# GPG Key Generation Script for Terraform Provider Publishing
# This script generates a GPG key pair for signing Terraform provider releases

set -e

echo "============================================"
echo "  GPG Key Generator for Terraform Registry  "
echo "============================================"
echo ""

# Get email from git config
EMAIL=$(git config user.email 2>/dev/null || echo "")
NAME=$(git config user.name 2>/dev/null || echo "Terraform Provider BCM")

if [ -z "$EMAIL" ]; then
    echo "Error: No email found in git config"
    echo "Please run: git config user.email 'your-email@example.com'"
    exit 1
fi

echo "Using email: $EMAIL"
echo "Using name: $NAME"
echo ""

# Generate random passphrase (32 characters)
PASSPHRASE=$(openssl rand -base64 32 | tr -d '\n')

# Output directory
OUTPUT_DIR="/workspace/.gpg-keys"
mkdir -p "$OUTPUT_DIR"
chmod 700 "$OUTPUT_DIR"

# Create a temporary GNUPGHOME to avoid agent issues
export GNUPGHOME=$(mktemp -d)
chmod 700 "$GNUPGHOME"

# Configure gpg-agent to use loopback pinentry
cat > "$GNUPGHOME/gpg-agent.conf" << EOF
allow-loopback-pinentry
EOF

cat > "$GNUPGHOME/gpg.conf" << EOF
pinentry-mode loopback
EOF

# Create GPG batch file for non-interactive key generation
GPG_BATCH_FILE=$(mktemp)
cat > "$GPG_BATCH_FILE" << EOF
%echo Generating GPG key for Terraform Provider BCM
Key-Type: RSA
Key-Length: 4096
Subkey-Type: RSA
Subkey-Length: 4096
Name-Real: $NAME
Name-Email: $EMAIL
Expire-Date: 0
Passphrase: $PASSPHRASE
%commit
%echo Key generation complete
EOF

echo "Generating GPG key pair (this may take a moment)..."
echo ""

# Kill any existing gpg-agent and start fresh
gpgconf --kill gpg-agent 2>/dev/null || true

# Generate the key
gpg --batch --gen-key "$GPG_BATCH_FILE" 2>&1

# Clean up batch file
rm -f "$GPG_BATCH_FILE"

# Get the key ID
KEY_ID=$(gpg --list-secret-keys --keyid-format=long 2>/dev/null | grep -E "^sec" | head -1 | sed 's/.*rsa4096\///' | sed 's/ .*//')

if [ -z "$KEY_ID" ]; then
    echo "Error: Failed to get GPG key ID"
    echo "Listing all secret keys:"
    gpg --list-secret-keys
    rm -rf "$GNUPGHOME"
    exit 1
fi

echo ""
echo "GPG Key generated successfully!"
echo "Key ID: $KEY_ID"
echo ""

# Export public key
PUBLIC_KEY_FILE="$OUTPUT_DIR/public-key.asc"
gpg --armor --export "$KEY_ID" > "$PUBLIC_KEY_FILE"
echo "Public key exported to: $PUBLIC_KEY_FILE"

# Export private key
PRIVATE_KEY_FILE="$OUTPUT_DIR/private-key.asc"
echo "$PASSPHRASE" | gpg --batch --yes --pinentry-mode loopback --passphrase-fd 0 --armor --export-secret-keys "$KEY_ID" > "$PRIVATE_KEY_FILE"
chmod 600 "$PRIVATE_KEY_FILE"
echo "Private key exported to: $PRIVATE_KEY_FILE"

# Clean up temporary GNUPGHOME
rm -rf "$GNUPGHOME"

echo ""
echo "============================================"
echo "              OUTPUT SUMMARY                "
echo "============================================"
echo ""
echo "GPG Key ID:"
echo "$KEY_ID"
echo ""
echo "PASSPHRASE (save this securely for GitHub Secrets):"
echo "$PASSPHRASE"
echo ""
echo "============================================"
echo "         GITHUB SECRETS TO ADD             "
echo "============================================"
echo ""
echo "Go to: Repository Settings > Secrets and variables > Actions"
echo ""
echo "1. Secret name: GPG_PRIVATE_KEY"
echo "   Value: (contents below)"
echo ""
cat "$PRIVATE_KEY_FILE"
echo ""
echo "2. Secret name: PASSPHRASE"
echo "   Value: $PASSPHRASE"
echo ""
echo "============================================"
echo "      PUBLIC KEY FOR TERRAFORM REGISTRY    "
echo "============================================"
echo ""
cat "$PUBLIC_KEY_FILE"
echo ""
echo "============================================"
echo "         NEXT STEPS                        "
echo "============================================"
echo ""
echo "1. Copy the public key above to Terraform Registry:"
echo "   https://registry.terraform.io > User Settings > Signing Keys"
echo ""
echo "2. Add GitHub secrets (GPG_PRIVATE_KEY and PASSPHRASE)"
echo ""
echo "3. Register provider at:"
echo "   https://registry.terraform.io > Publish > Provider"
echo "   Select: hashi-demo-lab/terraform-provider-bcm"
echo ""
echo "4. Create first release:"
echo "   git tag v0.1.0 && git push origin v0.1.0"
echo ""
