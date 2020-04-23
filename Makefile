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
	cd cmd/api && CONFIG_FILE=config-dev.json go run .

clearmq:
	docker exec -it bcd-mq rabbitmqctl stop_app
	docker exec -it bcd-mq rabbitmqctl reset
	docker exec -it bcd-mq rabbitmqctl start_app

local:
	docker-compose -f docker-compose.local.yml up -d --build

aliases:
	cd scripts/aliases && go run .

migration:
	cd scripts/migration && go run .

upd:
	docker-compose -f docker-compose.yml docker-compose.prod.yml up -d --build

es-aws:
	cd scripts/es-aws && go build .

s3-creds:
	docker exec -it bcd-elastic bin/elasticsearch-keystore add --stdin s3.client.default.access_key
	docker exec -it bcd-elastic bin/elasticsearch-keystore add --stdin s3.client.default.secret_key

s3-repo: es-aws
	./scripts/es-aws/es-aws -a create_repository

s3-restore: es-aws
	./scripts/es-aws/es-aws -a restore

s3-policy: es-aws
	./scripts/es-aws/es-aws -a set_policy
