# Define the version/tag for your Docker image
HASMPC_VERSION ?= latest

# Define the full image name
IMAGE_NAME = hasmcp/hasmcp-ce

# Combine them for the full reference
DOCKER_IMAGE = $(IMAGE_NAME):$(HASMPC_VERSION)

setup:
	mkdir _certs _storage; \
		chmod 0777 _certs; \
		chmod 0777 _storage; \
		wget https://raw.githubusercontent.com/hasmcp/hasmcp-ce/refs/heads/main/backend/cmd/server/.env.example -o .env;

push-build:
	docker buildx build --platform linux/amd64,linux/arm64 --tag hasmcp/hasmcp-ce:latest -f Dockerfile --push .

stats:
	docker stats hasmcp-ce

logs:
	docker logs -f hasmcp-ce

update:
	docker stop hasmcp-ce || true; \
	docker rm hasmcp-ce || true; \
	docker image prune -f; \
	docker pull $(DOCKER_IMAGE); \
	docker run --env-file .env -p 80:80 -p 443:443 --name hasmcp-ce -v ./_certs:/_certs -v ./_storage:/_storage -d --restart always $(DOCKER_IMAGE)

restart:
	docker stop hasmcp-ce; \
	docker rm hasmcp-ce; \
	docker run --env-file .env -p 80:80 -p 443:443 --name hasmcp-ce -v ./_certs:/_certs -v ./_storage:/_storage -d --restart always $(DOCKER_IMAGE)

build:
	docker build . -t hasmcp-ce

build-ui:
	rm -rf ./backend/cmd/server/public/*
	cd ./frontend && npm run build
	cp -R ./frontend/dist/* ./backend/cmd/server/public/

	# Compress files in the root directory (public/)
	find ./backend/cmd/server/public -maxdepth 1 -type f -not -name "*.gz" -not -name "*.br" -exec gzip -9 -f -k {} +
	find ./backend/cmd/server/public -maxdepth 1 -type f -not -name "*.gz" -not -name "*.br" -exec brotli -9 -f -k {} +

	# Compress files in the assets directory (public/assets/)
	find ./backend/cmd/server/public/assets -maxdepth 1 -type f -not -name "*.gz" -not -name "*.br" -exec gzip -9 -f -k {} +
	find ./backend/cmd/server/public/assets -maxdepth 1 -type f -not -name "*.gz" -not -name "*.br" -exec brotli -9 -f -k {} +
