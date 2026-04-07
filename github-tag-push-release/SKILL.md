---
name: github-tag-push-release
description: "Workflow for AnyWhereEtherNet: staging, committing, pushing to main, and creating a new version tag to trigger GitHub Actions release with a token from .env."
---

# GitHub Tag & Push Release Workflow

This skill automates the standard deployment workflow for the AnyWhereEtherNet project. It ensures that code changes are committed, pushed to the `main` branch with force, and a new version tag is created and pushed to trigger GitHub Actions' multi-platform release.

## Key Workflow Steps

### 1. Identify Version Increment
Determine the next version tag by checking the current latest tag:
`git describe --tags --abbrev=0`

### 2. Prepare Commit
- Stage all necessary changes.
- Use a descriptive commit message like `feat: <description>` or `fix: <description>`.

### 3. Retrieve Credentials
- Read the **raw GitHub Token** from the local `.env` file.
- Handle URL encoding for the username `daiwei7207@gmail.com` (using `%40`).

### 4. Deploy to Remote
- Target Repository: `https://github.com/bingbaga/AnyWhereEtherNet`
- **Push to Main**: Force push `HEAD` to the remote `main` branch.
- **Push Tag**: Create and push the version tag (e.g., `v0.3.5-f8`).

## Tool Usage Guide

Use the included script to perform the entire release process in one step:

```bash
bash scripts/release_push.sh "v0.3.5-f8" "Your commit message"
```

### Script Arguments:
1.  **Tag Name**: The version tag (e.g., `v0.3.5-f8`).
2.  **Commit Message**: Description of the changes.

## Best Practices
- **Verify Version**: Always check `version.go` or `git describe` to avoid tag collisions.
- **Clean Environment**: Ensure all test logs or sensitive config files are deleted or ignored before committing.
- **Monitor Actions**: After pushing, check the GitHub Actions tab in the repository to ensure the build and release are successful.
