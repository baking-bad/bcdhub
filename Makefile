-include .env
export $(shell sed 's/=.*//' .env)

.ONESHELL:

api:
	docker-compose up -d elastic mq db
	cd cmd/api && go run .

indexer:
	docker-compose up -d elastic mq
	cd cmd/indexer && go run .

metrics:
	docker-compose up -d elastic mq db
	cd cmd/metrics && go run .

compiler:
	docker-compose -f docker-compose.yml -f build/compiler/dev/docker-compose.yml up -d --build compiler-dev
	docker logs -f bcd-compiler-dev

seo:
ifeq ($(BCD_ENV), development)
	cd scripts/nginx && go run .
else
	docker-compose exec api seo
	docker-compose restart gui
endif

migration:
ifeq ($(BCD_ENV), development)
	cd scripts/migration && go run .
else
	docker-compose exec api migration
endif

rollback:
	cd scripts/esctl && go run . rollback -n $(NETWORK) -l $(LEVEL)

remove:
	cd scripts/esctl && go run . remove -n $(NETWORK) 

s3-creds:
	docker-compose exec elastic bash -c 'bin/elasticsearch-keystore add --force --stdin s3.client.default.access_key <<< "$$AWS_ACCESS_KEY_ID"'
	docker-compose exec elastic bash -c 'bin/elasticsearch-keystore add --force --stdin s3.client.default.secret_key <<< "$$AWS_SECRET_ACCESS_KEY"'
ifeq ($(BCD_ENV), development)
	cd scripts/esctl && go run . reload_secure_settings
else
	docker-compose exec api esctl reload_secure_settings
endif

s3-repo:
ifeq ($(BCD_ENV), development)
	cd scripts/esctl && go run . create_repository
else
	docker-compose exec api esctl create_repository
endif

s3-restore:
ifeq ($(BCD_ENV), development)
	cd scripts/esctl && go run . restore
else
	docker-compose exec api esctl restore
endif

s3-snapshot:
ifeq ($(BCD_ENV), development)
	cd scripts/esctl && go run . snapshot
else
	docker-compose exec api esctl snapshot
endif

s3-policy:
ifeq ($(BCD_ENV), development)
	cd scripts/esctl && go run . set_policy
else
	docker-compose exec api esctl set_policy
endif

es-reset:
	docker-compose rm -s -v -f elastic || true
	docker volume rm $$(docker volume ls -q | grep esdata | grep $$COMPOSE_PROJECT_NAME) || true
	docker-compose up -d elastic

clearmq:
	docker-compose exec mq rabbitmqctl stop_app
	docker-compose exec mq rabbitmqctl reset
	docker-compose exec mq rabbitmqctl start_app

test:
	go test ./...

lint:
	golangci-lint run

docs:
	# wget https://github.com/swaggo/swag/releases/download/v1.7.0/swag_1.7.0_Linux_x86_64.tar.gz
	# tar -zxvf swag_1.7.0_Linux_x86_64.tar.gz
	# sudo cp swag /usr/bin/swag
	cd cmd/api && swag init --parseDependency --parseInternal

images:
	docker-compose build

stable-images:
	TAG=$$(cat version.json | grep version | awk -F\" '{ print $$4 }' |  cut -d '.' -f1-2) docker-compose build

stable-pull:
	TAG=$$(cat version.json | grep version | awk -F\" '{ print $$4 }' |  cut -d '.' -f1-2) docker-compose pull

stable:
	TAG=$$(cat version.json | grep version | awk -F\" '{ print $$4 }' |  cut -d '.' -f1-2) docker-compose up -d

latest:
	docker-compose up -d

upgrade:
	$(MAKE) clearmq
	docker-compose down
	STABLE_TAG=$$(cat version.json | grep version | awk -F\" '{ print $$4 }' |  cut -d '.' -f1-2)
	TAG=$$STABLE_TAG $(MAKE) es-reset
	TAG=$$STABLE_TAG docker-compose up -d db mq api

restart:
	docker-compose restart api metrics indexer compiler

release:
	BCDHUB_VERSION=$$(cat version.json | grep version | awk -F\" '{ print $$4 }') && git tag $$BCDHUB_VERSION && git push origin $$BCDHUB_VERSION

db-dump:
	docker-compose exec db pg_dump -c bcd > dump_`date +%d-%m-%Y"_"%H_%M_%S`.sql

db-restore:
	docker-compose exec db psql --username $$POSTGRES_USER -v ON_ERROR_STOP=on bcd < $(BACKUP)

ps:
	docker ps --format "table {{.Names}}\t{{.RunningFor}}\t{{.Status}}\t{{.Ports}}"

sandbox-images:
	docker-compose -f docker-compose.sandbox.yml build

sandbox-up:
	COMPOSE_PROJECT_NAME=bcdbox docker-compose -f docker-compose.sandbox.yml up -d

sandbox-dev:
	COMPOSE_PROJECT_NAME=bcdbox BCD_ENV=you docker-compose -f docker-compose.sandbox.yml up -d elastic mq db api indexer metrics

sandbox-down:
	COMPOSE_PROJECT_NAME=bcdbox docker-compose -f docker-compose.sandbox.yml down

sandbox-clear:
	COMPOSE_PROJECT_NAME=bcdbox docker-compose -f docker-compose.sandbox.yml down -v
