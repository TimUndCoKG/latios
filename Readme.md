## Latios
Latios is a minimal reverse proxy with support for TLS utilizing letsencrypt certs on the host system. It can serve static files or reverse proxy traffic, even supporting websockets.
All routes are stored in a postgres database that is displayed next to it.

Latios can only redirect to services that are available in the docker network "latios-network"

#### Useful comamnds
1. docker pull ghcr.io/timundcokg/latios
2. docker compoes up -d --force-recreate
3. docker exec -it latios /bin/sh
4. curl -o compose.yml https://raw.githubusercontent.com/TimUndCoKG/latios/refs/heads/main/docker-compose.yml
