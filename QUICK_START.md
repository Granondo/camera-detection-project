# Quick Start Guide

## One-Time Setup on Home Server

### 1. Clone Repository
```bash
git clone https://github.com/YOUR_USERNAME/camera-detection-project.git
cd camera-detection-project
```

### 2. Create Environment File
```bash
cp .env.example .env
nano .env
```

Add your camera details:
```env
RTSP_URL=rtsp://192.168.1.71:554/stream1
CAMERA_USERNAME=your_email@gmail.com
CAMERA_PASSWORD=your_password
```

### 3. Install Nginx (Optional but Recommended)
```bash
sudo apt update && sudo apt install nginx -y
sudo cp nginx.conf /etc/nginx/sites-available/surveillance
sudo ln -s /etc/nginx/sites-available/surveillance /etc/nginx/sites-enabled/
sudo nginx -t && sudo systemctl restart nginx
```

### 4. Start Services
```bash
make start-full
```

**Access:**
- Frontend: `http://your-server-ip/` (or `http://your-server-ip:3000` without nginx)
- API: `http://your-server-ip/api/`

---

## CI/CD Setup (Auto-Deploy)

### Install GitHub Actions Runner

```bash
# 1. Download runner
mkdir ~/actions-runner && cd ~/actions-runner
curl -o actions-runner-linux-x64.tar.gz -L \
  https://github.com/actions/runner/releases/download/v2.311.0/actions-runner-linux-x64-2.311.0.tar.gz
tar xzf ./actions-runner-linux-x64.tar.gz

# 2. Configure (get token from GitHub: Settings → Actions → Runners)
./config.sh --url https://github.com/YOUR_USERNAME/camera-detection-project --token YOUR_TOKEN

# 3. Install as service
sudo ./svc.sh install
sudo ./svc.sh start
```

### Add GitHub Secrets

Go to: **Settings → Secrets and variables → Actions**

Add:
- `RTSP_URL`
- `CAMERA_USERNAME`
- `CAMERA_PASSWORD`

### Done!

Now every push to `main` branch will auto-deploy to your server.

---

## Daily Usage

### Check Status
```bash
make api-status       # API health
docker-compose ps     # All services
```

### View Logs
```bash
docker-compose logs -f api-server
docker-compose logs -f camera-detection
```

### Update Manually
```bash
cd ~/camera-detection-project
git pull
make stop-full
make start-full
```

### Cleanup
```bash
make clear-all        # Clear database + output
make output-clear     # Just clear recordings
docker system prune   # Clear old Docker images
```

---

## Useful Commands

| Command | Description |
|---------|-------------|
| `make start-full` | Start all services |
| `make stop-full` | Stop all services |
| `make api-status` | Check API status |
| `make api-events` | List recent events |
| `make logs-api` | View API logs |
| `make rabbitmq-ui` | Open RabbitMQ UI |
| `make dashboard` | Open frontend |

---

## Troubleshooting

### Services won't start
```bash
docker-compose down --remove-orphans
make start-full
```

### Out of memory
```bash
# Increase Docker memory in Docker Desktop settings
# Or build sequentially (already configured in Makefile)
```

### Port already in use
```bash
docker-compose down
sudo lsof -ti:8080 | xargs kill -9  # Replace 8080 with conflicting port
```

### Can't access from network
```bash
sudo ufw allow 80/tcp    # For nginx
sudo ufw allow 3000/tcp  # For direct access
```

---

## File Locations

- **Frontend**: `surveillance-next/`
- **API**: `cmd/api_service/`
- **Camera Service**: `cmd/server/`
- **Detection Worker**: `cmd/detection_worker/`
- **Config**: `.env`, `docker-compose.yml`
- **Output**: `output/` (videos, frames)
- **Logs**: `logs/`

---

## More Help

- **Deployment**: See `DEPLOYMENT.md`
- **CI/CD**: See `CI_CD_SETUP.md`
- **Docker Issues**: See `DOCKER_BUILD.md`
- **GitHub Issues**: https://github.com/YOUR_USERNAME/camera-detection-project/issues
