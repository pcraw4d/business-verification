# Docker Setup for Local Development

## Prerequisites

Before running the local microservices setup, you need Docker and Docker Compose installed and running.

## Installing Docker

### macOS

1. **Download Docker Desktop**:
   - Visit https://www.docker.com/products/docker-desktop
   - Download Docker Desktop for Mac
   - Install the .dmg file

2. **Start Docker Desktop**:
   - Open Docker Desktop from Applications
   - Wait for Docker to start (whale icon in menu bar should be steady)

3. **Verify Installation**:
   ```bash
   docker --version
   docker-compose --version
   docker ps
   ```

### Linux

```bash
# Install Docker
curl -fsSL https://get.docker.com -o get-docker.sh
sudo sh get-docker.sh

# Install Docker Compose
sudo curl -L "https://github.com/docker/compose/releases/latest/download/docker-compose-$(uname -s)-$(uname -m)" -o /usr/local/bin/docker-compose
sudo chmod +x /usr/local/bin/docker-compose

# Start Docker service
sudo systemctl start docker
sudo systemctl enable docker

# Verify
docker --version
docker-compose --version
```

## Starting Docker

### macOS (Docker Desktop)

1. Open Docker Desktop from Applications
2. Wait for the Docker icon in the menu bar to show "Docker Desktop is running"
3. You can also start it from Terminal:
   ```bash
   open -a Docker
   ```

### Linux

```bash
sudo systemctl start docker
```

## Verifying Docker is Running

```bash
# Check Docker daemon
docker ps

# If you see an error about the Docker daemon, Docker is not running
# If you see a table (even if empty), Docker is running
```

## Troubleshooting

### "Cannot connect to the Docker daemon"

**Solution**: Start Docker Desktop (macOS) or Docker service (Linux)

### "Permission denied" (Linux)

**Solution**: Add your user to the docker group:
```bash
sudo usermod -aG docker $USER
# Log out and log back in for changes to take effect
```

### Docker Desktop won't start (macOS)

1. Check System Requirements:
   - macOS 10.15 or newer
   - At least 4GB RAM
   - VirtualBox prior to version 4.3.30 must NOT be installed

2. Reset Docker Desktop:
   - Docker Desktop menu → Troubleshoot → Reset to factory defaults

## Next Steps

Once Docker is running:

1. **Create .env.local** (if not already done):
   ```bash
   cp railway.env .env.local
   # Edit .env.local and set ENV=local, LOG_LEVEL=debug
   ```

2. **Start services**:
   ```bash
   make start-local
   ```

3. **Check status**:
   ```bash
   make status-local
   ```

4. **View logs**:
   ```bash
   make logs-local
   ```

## Alternative: Use Unified Server (No Docker Required)

If Docker is not available, you can use the unified server mode:

```bash
# Start unified server (single binary, no Docker)
make start-unified
```

This runs all functionality in a single process, useful for quick development but doesn't match the production microservices architecture.

