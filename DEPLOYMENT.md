# Deployment Guide for Linux Home Server

This guide will help you deploy the Camera Detection Project on your Linux home server.

## Prerequisites

- Linux server (Ubuntu 20.04+ recommended)
- Docker and Docker Compose installed
- Nginx installed (optional but recommended)
- At least 6GB RAM for Docker
- Camera accessible on your network

## Option 1: Direct Access (Simple Setup)

### 1. Copy Project to Server

```bash
# On your local machine
scp -r camera-detection-project user@your-server:/home/user/

# Or clone from git
ssh user@your-server
git clone <your-repo-url> camera-detection-project
cd camera-detection-project
```

### 2. Configure Environment

```bash
# Create .env file with your camera settings
cp .env.example .env
nano .env
```

Update with your camera IP and credentials:
```env
RTSP_URL=rtsp://192.168.1.71:554/stream1
CAMERA_USERNAME=your_email@gmail.com
CAMERA_PASSWORD=your_password
```

### 3. Start Services

```bash
make start-full
```

### 4. Access Services

- Frontend: `http://your-server-ip:3000`
- API: `http://your-server-ip:8080`
- Kibana: `http://your-server-ip:5601`
- RabbitMQ: `http://your-server-ip:15672`

## Option 2: Nginx Reverse Proxy (Recommended)

This setup provides cleaner URLs and better security.

### 1. Install Nginx

```bash
sudo apt update
sudo apt install nginx -y
```

### 2. Configure Nginx

```bash
# Copy nginx config
sudo cp nginx.conf /etc/nginx/sites-available/surveillance

# Create symbolic link
sudo ln -s /etc/nginx/sites-available/surveillance /etc/nginx/sites-enabled/

# Remove default site (optional)
sudo rm /etc/nginx/sites-enabled/default

# Test configuration
sudo nginx -t

# Restart nginx
sudo systemctl restart nginx
sudo systemctl enable nginx
```

### 3. Start Services

```bash
make start-full
```

### 4. Access Services

Now you can access everything through nginx on port 80:

- Frontend: `http://your-server-ip/`
- API: `http://your-server-ip/api/`
- Kibana: `http://your-server-ip/kibana/`
- RabbitMQ: `http://your-server-ip/rabbitmq/`

## Option 3: HTTPS with SSL (Production)

For secure access with HTTPS:

### 1. Get a Domain Name

Point a domain to your server's IP (e.g., `surveillance.yourdomain.com`)

### 2. Install Certbot

```bash
sudo apt install certbot python3-certbot-nginx -y
```

### 3. Get SSL Certificate

```bash
# Update nginx config with your domain first
sudo nano /etc/nginx/sites-available/surveillance
# Change: server_name surveillance.yourdomain.com;

# Get certificate
sudo certbot --nginx -d surveillance.yourdomain.com

# Auto-renewal is set up automatically
```

### 4. Access with HTTPS

- `https://surveillance.yourdomain.com/`

## Firewall Configuration

### Allow Required Ports

```bash
# If using direct access
sudo ufw allow 3000/tcp  # Frontend
sudo ufw allow 8080/tcp  # API
sudo ufw allow 5601/tcp  # Kibana

# If using nginx
sudo ufw allow 80/tcp    # HTTP
sudo ufw allow 443/tcp   # HTTPS (if using SSL)

# SSH (important!)
sudo ufw allow 22/tcp

# Enable firewall
sudo ufw enable
```

## Auto-Start on Boot

Make services start automatically when server reboots:

### 1. Create Systemd Service

```bash
sudo nano /etc/systemd/system/surveillance.service
```

Add this content:

```ini
[Unit]
Description=Camera Detection Surveillance System
Requires=docker.service
After=docker.service

[Service]
Type=oneshot
RemainAfterExit=yes
WorkingDirectory=/home/user/camera-detection-project
ExecStart=/usr/bin/docker-compose up -d
ExecStop=/usr/bin/docker-compose down
User=user
Group=docker

[Install]
WantedBy=multi-user.target
```

### 2. Enable Service

```bash
sudo systemctl daemon-reload
sudo systemctl enable surveillance.service
sudo systemctl start surveillance.service
```

## Remote Access (Optional)

### Access from Outside Your Home Network

#### Option A: Port Forwarding
1. Log into your router
2. Forward port 80 (or 443) to your server's local IP
3. Access via your public IP: `http://your-public-ip/`

#### Option B: VPN (More Secure)
Set up WireGuard or OpenVPN to securely access your home network

#### Option C: Cloudflare Tunnel (Easy & Secure)
Use Cloudflare Tunnel to expose your service without port forwarding

## Monitoring and Maintenance

### Check Service Status

```bash
# Docker services
docker-compose ps

# Nginx status
sudo systemctl status nginx

# View logs
docker-compose logs -f api-server
docker-compose logs -f camera-detection
```

### Update Project

```bash
cd camera-detection-project
git pull
make stop-full
make start-full
```

### Backup Database

```bash
# Backup PostgreSQL
docker-compose exec postgres pg_dump -U postgres surveillance > backup.sql

# Restore
cat backup.sql | docker-compose exec -T postgres psql -U postgres surveillance
```

## Performance Tips

### 1. Increase Docker Memory
```bash
# Edit /etc/docker/daemon.json
sudo nano /etc/docker/daemon.json
```

Add:
```json
{
  "default-address-pools": [
    {"base":"172.17.0.0/16","size":24}
  ],
  "log-driver": "json-file",
  "log-opts": {
    "max-size": "10m",
    "max-file": "3"
  }
}
```

```bash
sudo systemctl restart docker
```

### 2. Enable Log Rotation

Already configured in docker-compose.yml for all services.

### 3. Regular Cleanup

```bash
# Clean old Docker images/containers
docker system prune -a --volumes -f

# Clear old recordings (optional)
find /path/to/output -name "*.mp4" -mtime +7 -delete
```

## Troubleshooting

### Services Won't Start
```bash
# Check Docker
sudo systemctl status docker
docker-compose logs

# Check ports
sudo netstat -tulpn | grep LISTEN
```

### Nginx Errors
```bash
# Check nginx config
sudo nginx -t

# View error logs
sudo tail -f /var/log/nginx/error.log
```

### Can't Access from Network
```bash
# Check firewall
sudo ufw status

# Check if services are listening
sudo netstat -tulpn | grep LISTEN
```

### Out of Disk Space
```bash
# Check disk usage
df -h

# Clean Docker
docker system prune -a --volumes -f

# Clean old recordings
make output-clear
```

## Security Best Practices

1. **Change Default Passwords**: Update PostgreSQL, RabbitMQ, Redis passwords in docker-compose.yml
2. **Use Firewall**: Only open necessary ports
3. **Enable HTTPS**: Use SSL certificates for production
4. **Regular Updates**: Keep system and Docker images updated
5. **Backup Regularly**: Backup database and important data
6. **Monitor Logs**: Check logs for suspicious activity
7. **Use VPN**: For remote access instead of exposing ports

## Recommended Server Specs

- **Minimum**: 4 CPU cores, 8GB RAM, 100GB disk
- **Recommended**: 8 CPU cores, 16GB RAM, 500GB SSD
- **Network**: 1 Gbps for reliable video streaming

## CI/CD Pipeline

Automate deployment when pushing to the `main` branch on GitHub.

### Quick Setup (Self-Hosted Runner - Recommended)

1. **Install GitHub Actions Runner on your server:**
   ```bash
   cd ~/
   mkdir actions-runner && cd actions-runner

   # For x64
   curl -o actions-runner-linux-x64.tar.gz -L \
     https://github.com/actions/runner/releases/download/v2.311.0/actions-runner-linux-x64-2.311.0.tar.gz
   tar xzf ./actions-runner-linux-x64.tar.gz

   # For ARM64 (Raspberry Pi)
   # curl -o actions-runner-linux-arm64.tar.gz -L \
   #   https://github.com/actions/runner/releases/download/v2.311.0/actions-runner-linux-arm64-2.311.0.tar.gz
   # tar xzf ./actions-runner-linux-arm64.tar.gz
   ```

2. **Configure runner:**
   - Go to GitHub repo: **Settings → Actions → Runners → New self-hosted runner**
   - Copy and run the configuration command shown
   - Install as service: `sudo ./svc.sh install && sudo ./svc.sh start`

3. **Add GitHub Secrets:**
   Go to: **Settings → Secrets and variables → Actions → New repository secret**

   Add these secrets:
   - `RTSP_URL` - Your camera RTSP URL
   - `CAMERA_USERNAME` - Camera username
   - `CAMERA_PASSWORD` - Camera password

4. **Push to main branch:**
   ```bash
   git push origin main
   ```

   Watch deployment in the **Actions** tab!

### Features

- ✅ Auto-deploy on push to `main`
- ✅ Manual deployment trigger
- ✅ Health checks after deployment
- ✅ Automatic cleanup of old Docker images
- ✅ Test workflow for pull requests

### Complete Guide

See **CI_CD_SETUP.md** for:
- Detailed setup instructions
- SSH deployment alternative
- Troubleshooting
- Rollback procedures
- Advanced configurations

## Questions?

Check the main README.md, DOCKER_BUILD.md, or CI_CD_SETUP.md for more information.
