# nE Production Operations & Deployment Guide

## 1. Hardware Sizing Matrix

| Metric / Scale | Tier 1: Small (< 10k tracks) | Tier 2: Medium (10k - 100k tracks) | Tier 3: Large (100k - 500k+ tracks) |
| :--- | :--- | :--- | :--- |
| **CPU** | 1-2 vCPU (x86_64 or ARM64) | 2-4 vCPU | 4-8 vCPU |
| **RAM** | 256 MB - 512 MB | 512 MB - 1 GB | 2 GB - 4 GB |
| **Storage (DB & Cache)** | 5 GB SSD/NVMe | 25 GB SSD/NVMe | 100 GB+ SSD/NVMe |
| **Concurrent Streams** | 1 - 3 streams | 5 - 15 streams | 25+ streams |
| **Transcoding Limit** | 1 - 2 workers | 4 workers | 8+ workers |

---

## 2. Production Deployment (Docker + Caddy)

1. Clone or place the repository on your server.
2. Edit `deploy/Caddyfile` with your domain name (`music.yourdomain.com`) and contact email.
3. Edit `deploy/docker-compose.prod.yml` to point `/mnt/storage/music` to your host music collection mount.
4. Launch the services:
   ```bash
   cd deploy
   docker compose -f docker-compose.prod.yml up -d --build
   ```
5. Verify health:
   ```bash
   docker compose -f docker-compose.prod.yml ps
   curl -I https://music.yourdomain.com/health
   ```

---

## 3. Database Maintenance, Backup & Disaster Recovery

### Hot Online Backup
nE executes non-blocking atomic backups using SQLite `VACUUM INTO`:
```bash
# Direct CLI hot backup
ne backup /backups/ne_snapshot_$(date +%Y%m%d_%H%M%S).db

# Or inside Docker container
docker exec -t ne-audio-server ne backup /data/backup_$(date +%Y%m%d).db
```

### Integrity Verification
```bash
docker exec -t ne-audio-server ne check-db
```

### Search Index Rebuilding
```bash
docker exec -t ne-audio-server ne rebuild-fts
```

### Statistics Recalculation
```bash
docker exec -t ne-audio-server ne recalculate-stats
```

### Disaster Recovery / Database Restoration
1. Stop the nE server:
   ```bash
   docker compose -f docker-compose.prod.yml stop ne
   ```
2. Replace `/data/ne.db` with the verified backup `.db` file:
   ```bash
   cp /backups/ne_snapshot_20260830.db /var/lib/docker/volumes/deploy_ne_data/_data/ne.db
   rm -f /var/lib/docker/volumes/deploy_ne_data/_data/ne.db-wal
   rm -f /var/lib/docker/volumes/deploy_ne_data/_data/ne.db-shm
   ```
3. Restart the container:
   ```bash
   docker compose -f docker-compose.prod.yml start ne
   ```
