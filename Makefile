build-docker:
	docker build -t ghcr.io/timundcokg/latios:latest .

build-frontend:
	pnpm -C latios-frontend run build

recreate:
	sudo docker compose up --force-recreate

local:
	make build-frontend
	make build-docker
	make recreate