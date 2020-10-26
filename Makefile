include .env
export $(shell sed 's/=.*//' .env)

api:
	docker-compose up -d $$COMPOSE_PREFIX-elastic $$COMPOSE_PREFIX-mq $$COMPOSE_PREFIX-db
	cd cmd/api && go run .

indexer:
	docker-compose up -d $$COMPOSE_PREFIX-elastic $$COMPOSE_PREFIX-mq $$COMPOSE_PREFIX-db
	cd cmd/indexer && go run .

metrics:
	docker-compose up -d $$COMPOSE_PREFIX-elastic $$COMPOSE_PREFIX-mq $$COMPOSE_PREFIX-db
	cd cmd/metrics && go run .

compiler:
	docker-compose -f docker-compose.yml -f build/compiler/dev/docker-compose.yml up -d --build compiler-dev
	docker logs -f bcd-compiler-dev

sitemap:
	cd scripts/sitemap && go run .

migration:
	cd scripts/migration && go run .

rollback:
	cd scripts/esctl && go run . rollback -n $(NETWORK) -l $(LEVEL)

remove:
	cd scripts/esctl && go run . remove -n $(NETWORK) 

s3-creds:
	docker exec -it $$COMPOSE_PREFIX-elastic bash -c 'bin/elasticsearch-keystore add --force --stdin s3.client.default.access_key <<< "$$AWS_ACCESS_KEY_ID"'
	docker exec -it $$COMPOSE_PREFIX-elastic bash -c 'bin/elasticsearch-keystore add --force --stdin s3.client.default.secret_key <<< "$$AWS_SECRET_ACCESS_KEY"'
	cd scripts/esctl && go run . reload_secure_settings

s3-repo:
	cd scripts/esctl && go run . create_repository

s3-restore:
	cd scripts/esctl && go run . restore

s3-snapshot:
	cd scripts/esctl && go run . snapshot

s3-policy:
	cd scripts/esctl && go run . set_policy

es-reset:
	docker stop $$COMPOSE_PREFIX-elastic || true
	docker rm $$COMPOSE_PREFIX-elastic || true
	docker volume rm bcdhub_esdata || true
	docker-compose up -d $$COMPOSE_PREFIX-elastic

clearmq:
	docker exec -it $$COMPOSE_PREFIX-mq rabbitmqctl stop_app
	docker exec -it $$COMPOSE_PREFIX-mq rabbitmqctl reset
	docker exec -it $$COMPOSE_PREFIX-mq rabbitmqctl start_app

test:
	go test ./...
  
docs:
	# wget https://github.com/swaggo/swag/releases/download/v1.6.6/swag_1.6.6_Linux_x86_64.tar.gz
	# tar -zxvf swag_1.6.6_Linux_x86_64.tar.gz
	# sudo cp swag /usr/bin/swag
	cd cmd/api && swag init --parseDependency

images:
	docker-compose build

stable-images:
	TAG=$$STABLE_TAG docker-compose build

stable-pull:
	TAG=$$STABLE_TAG docker-compose pull

stable:
	TAG=$$STABLE_TAG docker-compose up -d

upgrade:
	$(MAKE) clearmq
	docker-compose down
	TAG=$$STABLE_TAG $(MAKE) es-reset
	docker-compose up -d $$COMPOSE_PREFIX-db $$COMPOSE_PREFIX-mq

restart:
	docker-compose restart $$COMPOSE_PREFIX-api $$COMPOSE_PREFIX-metrics $$COMPOSE_PREFIX-indexer $$COMPOSE_PREFIX-compiler

release:
	BCDHUB_VERSION=$$(cat version.json | grep version | awk -F\" '{ print $$4 }')
	git tag $$BCDHUB_VERSION && git push origin $$BCDHUB_VERSION

db-dump:
	docker exec -it $$COMPOSE_PREFIX-db pg_dump -c bcd > dump_`date +%d-%m-%Y"_"%H_%M_%S`.sql

db-restore:
	docker exec -i $$COMPOSE_PREFIX-db psql --username $$POSTGRES_USER -v ON_ERROR_STOP=on bcd < $(BACKUP)
