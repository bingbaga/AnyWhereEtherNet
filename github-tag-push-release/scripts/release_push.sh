#!/bin/bash
# scripts/release_push.sh
# Automates the tag and push process for AnyWhereEtherNet releases.

set -euo pipefail

TARGET_REPO="https://github.com/bingbaga/AnyWhereEtherNet"
USERNAME="daiwei7207@gmail.com"
BRANCH="main"
ENV_FILE=".env"

# 1. Get arguments
TAG_NAME=$1
COMMIT_MSG=$2

if [ -z "$TAG_NAME" ] || [ -z "$COMMIT_MSG" ]; then
    echo "Usage: $0 <tag_name> <commit_message>"
    exit 1
fi

# 2. Commit changes
git add .
git commit -m "$COMMIT_MSG" || echo "No changes to commit"

# 3. Read token and encode username
if [ ! -f "$ENV_FILE" ]; then
    echo "Error: $ENV_FILE not found."
    exit 1
fi
TOKEN=$(cat "$ENV_FILE")
ENCODED_USER=$(echo "$USERNAME" | sed 's/@/%40/g')
REMOTE_URL="https://${ENCODED_USER}:${TOKEN}@$(echo $TARGET_REPO | sed 's/https:\/\///')"

# 4. Push to main branch
echo "Pushing to $BRANCH..."
git push "$REMOTE_URL" HEAD:"$BRANCH" --force

# 5. Create and push tag
echo "Tagging $TAG_NAME..."
git tag "$TAG_NAME" || git tag -f "$TAG_NAME"
git push "$REMOTE_URL" "$TAG_NAME" --force

echo "Successfully pushed and tagged $TAG_NAME. GitHub Action should be triggered."
