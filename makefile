build:
	GOOS=linux go build -o app main.go
	docker build -f Dockerfile -t app .
up:
	docker-compose -f ./docker-compose/docker-compose.yaml -p "apm-demo" up -d

.PHONY: app up