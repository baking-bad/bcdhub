include .env
export $(shell sed 's/=.*//' .env)

deploy: export TAG=$(shell git pull -q && git describe --abbrev=0 --tags)
deploy:
	git pull
	docker-compose pull
	docker-compose -f docker-compose.yml -f docker-compose.prod.yml up -d
	docker ps

api:
	cd cmd/api && CONFIG_FILE=config-dev.json go run .

indexer:
	cd cmd/indexer && CONFIG_FILE=config-dev.json go run .

metrics:
	cd cmd/metrics && CONFIG_FILE=config-dev.json go run .

clearmq:
	docker exec -it bcd-mq rabbitmqctl stop_app
	docker exec -it bcd-mq rabbitmqctl reset
	docker exec -it bcd-mq rabbitmqctl start_app

aliases:
	cd scripts/aliases && go run .

migration:
	cd scripts/migration && go run .

upd:
	docker-compose -f docker-compose.yml docker-compose.prod.yml up -d --build

s3-creds:
	docker exec -it bcd-elastic bin/elasticsearch-keystore add s3.client.default.access_key
	docker exec -it bcd-elastic bin/elasticsearch-keystore add s3.client.default.secret_key