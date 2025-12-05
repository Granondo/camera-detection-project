# CI/CD Setup Guide

Automatic deployment to your home server when you push to the `main` branch on GitHub.

## Overview

We provide two CI/CD options:

1. **Self-Hosted Runner** (Recommended) - Runner runs on your home server
2. **SSH Deployment** - GitHub-hosted runner deploys via SSH

## Option 1: Self-Hosted Runner (Recommended)

This is the best option for home servers as it doesn't require exposing SSH or managing keys.

### Benefits
- ✅ No SSH exposure needed
- ✅ Faster builds (local network)
- ✅ More secure (everything stays on your network)
- ✅ Free for private repos

### Setup Steps

#### 1. Install Runner on Your Server

```bash
# SSH into your server
ssh user@your-server

# Create runner directory
mkdir -p ~/actions-runner && cd ~/actions-runner

# Download the latest runner (for Linux x64)
curl -o actions-runner-linux-x64.tar.gz -L \
  https://github.com/actions/runner/releases/download/v2.311.0/actions-runner-linux-x64-2.311.0.tar.gz

# Extract
tar xzf ./actions-runner-linux-x64.tar.gz

# For ARM64 (Raspberry Pi, etc), use:
# curl -o actions-runner-linux-arm64.tar.gz -L \
#   https://github.com/actions/runner/releases/download/v2.311.0/actions-runner-linux-arm64-2.311.0.tar.gz
# tar xzf ./actions-runner-linux-arm64.tar.gz
```

#### 2. Configure Runner

Go to your GitHub repository:
1. Click **Settings** → **Actions** → **Runners**
2. Click **New self-hosted runner**
3. Select **Linux** and your architecture
4. Copy the configuration command shown

```bash
# Run the configuration command from GitHub (it will look like this):
./config.sh --url https://github.com/YOUR_USERNAME/camera-detection-project \
  --token YOUR_TOKEN

# When prompted:
# - Runner group: Default
# - Runner name: home-server (or whatever you prefer)
# - Work folder: _work (default)
# - Run as a service: yes
```

#### 3. Install and Start Runner as Service

```bash
# Install service
sudo ./svc.sh install

# Start service
sudo ./svc.sh start

# Check status
sudo ./svc.sh status
```

#### 4. Configure GitHub Secrets

Go to your repository: **Settings** → **Secrets and variables** → **Actions**

Click **New repository secret** and add:

| Secret Name | Value | Example |
|------------|-------|---------|
| `RTSP_URL` | Your camera RTSP URL | `rtsp://192.168.1.71:554/stream1` |
| `CAMERA_USERNAME` | Camera username | `your_email@gmail.com` |
| `CAMERA_PASSWORD` | Camera password | `your_password` |

#### 5. Test Deployment

```bash
# Make a small change and push to main
git add .
git commit -m "Test CI/CD deployment"
git push origin main

# Watch the workflow run on GitHub:
# Go to: Actions tab in your repository
```

### Verify Runner is Working

```bash
# On your server, check runner logs
journalctl -u actions.runner.* -f

# Or check the runner service
sudo ./svc.sh status
```

### Troubleshooting Self-Hosted Runner

**Runner not appearing online:**
```bash
# Check service status
sudo ./svc.sh status

# Check logs
journalctl -u actions.runner.* -n 50

# Restart service
sudo ./svc.sh stop
sudo ./svc.sh start
```

**Permission issues:**
```bash
# Make sure runner user has Docker access
sudo usermod -aG docker $USER

# You may need to restart the runner service
sudo ./svc.sh restart
```

## Option 2: SSH Deployment (Alternative)

Use this if you prefer GitHub-hosted runners or can't run a self-hosted runner.

### Setup Steps

#### 1. Generate SSH Key on Your Server

```bash
# On your server
ssh-keygen -t ed25519 -C "github-actions" -f ~/.ssh/github_deploy

# Add public key to authorized_keys
cat ~/.ssh/github_deploy.pub >> ~/.ssh/authorized_keys

# Display private key (copy this)
cat ~/.ssh/github_deploy
```

#### 2. Configure GitHub Secrets

Add these secrets in GitHub: **Settings** → **Secrets and variables** → **Actions**

| Secret Name | Value | Example |
|------------|-------|---------|
| `SERVER_HOST` | Your server IP or hostname | `192.168.1.100` |
| `SERVER_USER` | SSH username | `ubuntu` |
| `SSH_PRIVATE_KEY` | Contents of `~/.ssh/github_deploy` | (entire private key) |
| `RTSP_URL` | Your camera RTSP URL | `rtsp://192.168.1.71:554/stream1` |
| `CAMERA_USERNAME` | Camera username | `your_email@gmail.com` |
| `CAMERA_PASSWORD` | Camera password | `your_password` |

#### 3. Activate SSH Workflow

```bash
# Rename the example file
mv .github/workflows/deploy-ssh.yml.example .github/workflows/deploy-ssh.yml

# Disable self-hosted workflow
mv .github/workflows/deploy.yml .github/workflows/deploy.yml.disabled

# Commit and push
git add .
git commit -m "Switch to SSH deployment"
git push origin main
```

#### 4. Ensure Server Accepts SSH from GitHub

```bash
# On your server, whitelist GitHub Actions IPs (optional)
# Or just allow SSH from anywhere (less secure)
sudo ufw allow 22/tcp
```

### Troubleshooting SSH Deployment

**SSH connection fails:**
- Verify SSH key is correct in GitHub secrets
- Check `SERVER_HOST` and `SERVER_USER` are correct
- Ensure port 22 is open: `sudo ufw status`
- Test SSH manually: `ssh -i ~/.ssh/github_deploy user@server`

**Permission denied:**
- Make sure public key is in `~/.ssh/authorized_keys`
- Check file permissions: `chmod 600 ~/.ssh/authorized_keys`

## Workflow Features

Both workflows include:

- ✅ Automatic deployment on push to `main`
- ✅ Manual deployment trigger (workflow_dispatch)
- ✅ Health checks after deployment
- ✅ Service status reporting
- ✅ Automatic Docker image cleanup
- ✅ Environment file management

## Testing Workflow

Separate test workflow runs on pull requests:

- Go code formatting check
- Go tests
- Docker build tests
- Frontend TypeScript compilation
- Linting

## Manual Deployment Trigger

You can manually trigger deployment:

1. Go to **Actions** tab in GitHub
2. Select **Deploy to Home Server**
3. Click **Run workflow**
4. Select `main` branch
5. Click **Run workflow**

## Monitoring Deployments

### View Deployment Logs in GitHub

1. Go to **Actions** tab
2. Click on the latest workflow run
3. Click on the **deploy** job
4. View step-by-step logs

### View Logs on Server

```bash
# For self-hosted runner
journalctl -u actions.runner.* -f

# Docker logs
docker-compose logs -f api-server
docker-compose logs -f camera-detection

# Nginx logs
sudo tail -f /var/log/nginx/access.log
sudo tail -f /var/log/nginx/error.log
```

## Rollback

If a deployment fails:

```bash
# SSH to server
ssh user@your-server
cd ~/camera-detection-project

# Check git history
git log --oneline

# Rollback to previous commit
git reset --hard COMMIT_HASH

# Restart services
make stop-full
make start-full
```

## Best Practices

1. **Test Locally First**: Always test changes locally before pushing
2. **Use Pull Requests**: Create PRs for review before merging to main
3. **Monitor First Deploy**: Watch the first deployment carefully
4. **Keep Secrets Updated**: Update secrets if camera/server credentials change
5. **Regular Backups**: Backup database before major updates
6. **Check Logs**: Monitor logs after each deployment

## Security Considerations

### Self-Hosted Runner
- Runner has access to your server
- Only use for private repositories you trust
- Keep runner software updated
- Run runner under non-root user with docker group access

### SSH Deployment
- SSH key only has access to deployment user
- Consider using dedicated deployment user with limited permissions
- Restrict SSH to GitHub Actions IPs if possible
- Use SSH key with passphrase for extra security

## Deployment Notifications (Optional)

### Slack Notifications

Add to your workflow:

```yaml
- name: Slack Notification
  if: always()
  uses: 8398a7/action-slack@v3
  with:
    status: ${{ job.status }}
    text: 'Deployment to home server'
    webhook_url: ${{ secrets.SLACK_WEBHOOK }}
```

### Discord Notifications

```yaml
- name: Discord Notification
  if: always()
  uses: sarisia/actions-status-discord@v1
  with:
    webhook: ${{ secrets.DISCORD_WEBHOOK }}
    status: ${{ job.status }}
    title: "Deployment"
```

## Advanced: Multi-Environment Setup

If you want staging + production:

1. Create branches: `staging` and `main`
2. Modify workflow to deploy based on branch:

```yaml
on:
  push:
    branches: [ main, staging ]

jobs:
  deploy:
    runs-on: self-hosted
    environment: ${{ github.ref_name }}  # main or staging
```

3. Configure different secrets for each environment in GitHub

## Cost

- **Self-Hosted Runner**: Free (uses your server resources)
- **SSH Deployment**: Free for public repos, includes 2000 minutes/month for private repos
- **GitHub Actions storage**: 500MB free

## Questions?

See DEPLOYMENT.md for server setup or open an issue on GitHub.
