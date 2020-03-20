include .env
export $(shell sed 's/=.*//' .env)

up:
	docker-compose up -d

build:
	docker-compose build

deploy: export TAG=$(shell git pull -q && git describe --abbrev=0 --tags)
deploy:
	git pull
	docker-compose pull
	docker-compose -f docker-compose.yml -f docker-compose.prod.yml up -d
	docker ps

api:
	cd cmd/api && go run .

clearmq:
	docker exec -it bcd-mq rabbitmqctl stop_app
	docker exec -it bcd-mq rabbitmqctl reset
	docker exec -it bcd-mq rabbitmqctl start_app

local:
	docker-compose -f docker-compose.local.yml up -d --build
