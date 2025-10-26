build-docker:
	docker build -t ghcr.io/timundcokg/latios:latest .

recreate:
	sudo docker compose up --force-recreate

local:
	make build-docker
	make recreate