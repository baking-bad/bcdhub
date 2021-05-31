-include .env
export $(shell sed 's/=.*//' .env)

LATEST_DUMP=/tmp/dump_latest.gz
BACKUP?=dump_latest.gz

.ONESHELL:

api:
	docker-compose up -d elastic mq db
	cd cmd/api && go run .

graphql:
	docker-compose up -d elastic mq db
	cd cmd/graphql && go run .

api-tester:
	docker-compose up -d elastic mq db
	cd scripts/api_tester && go run .

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
ifeq ($(BCD_ENV), development)
	cd scripts/bcdctl && go run . rollback -n $(NETWORK) -l $(LEVEL)
else
	docker-compose exec api bcdctl rollback -n $(NETWORK) -l $(LEVEL)
endif

s3-creds:
	docker-compose exec elastic bash -c 'bin/elasticsearch-keystore add --force --stdin s3.client.default.access_key <<< "$$AWS_ACCESS_KEY_ID"'
	docker-compose exec elastic bash -c 'bin/elasticsearch-keystore add --force --stdin s3.client.default.secret_key <<< "$$AWS_SECRET_ACCESS_KEY"'
ifeq ($(BCD_ENV), development)
	cd scripts/bcdctl && go run . reload_secure_settings
else
	docker-compose exec api bcdctl reload_secure_settings
endif

s3-repo:
ifeq ($(BCD_ENV), development)
	cd scripts/bcdctl && go run . create_repository
else
	docker-compose exec api bcdctl create_repository
endif

s3-restore:
	echo "Database restore..."
ifeq (,$(wildcard $(LATEST_DUMP)))
	aws s3 cp --profile bcd s3://bcd-db-snaps/$(BACKUP) $(LATEST_DUMP)
endif

	docker-compose exec -T db dropdb -U $(POSTGRES_USER) --if-exists indexer
	gunzip -dc $(LATEST_DUMP) | docker-compose exec -T db psql -U $(POSTGRES_USER) -v ON_ERROR_STOP=on bcd
	rm $(LATEST_DUMP)

	echo "Elasticsearch restore..."
ifeq ($(BCD_ENV), development)
	cd scripts/bcdctl && go run . restore
else
	docker-compose exec api bcdctl restore
endif

	echo "Contracts restore..."
	aws s3 cp --profile bcd s3://bcd-db-snaps/contracts.tar.gz /tmp/contracts.tar.gz
	rm -rf $(SHARE_PATH)/contracts/
	mkdir $(SHARE_PATH)/contracts/
	tar -C $(SHARE_PATH)/contracts/ -xzf /tmp/contracts.tar.gz

s3-snapshot:
	echo "Database snapshot..."
	docker-compose exec db pg_dump indexer --create -U $(POSTGRES_USER) | gzip -c > $(LATEST_DUMP)	
	aws s3 mv --profile bcd $(LATEST_DUMP) s3://bcd-db-snaps/dump_latest.gz

	echo "Elasticsearch snapshot..."
ifeq ($(BCD_ENV), development)
	cd scripts/bcdctl && go run . snapshot
else
	docker-compose exec api bcdctl snapshot
endif

	echo "Packing contracts..."
	cd $(SHARE_PATH)/contracts
	tar -zcvf /tmp/contracts.tar.gz .
	aws s3 mv --profile bcd /tmp/contracts.tar.gz s3://bcd-db-snaps/contracts.tar.gz
	rm /tmp/contracts.tar.gz

s3-list:
	echo "Database snapshots"
	aws s3 ls --profile bcd s3://bcd-db-snaps

	echo "Elasticsearch snapshots"
	aws s3 ls --profile bcd s3://bcd-elastic-snapshots

es-reset:
	docker-compose rm -s -v -f elastic || true
	docker volume rm $$(docker volume ls -q | grep esdata | grep $$COMPOSE_PROJECT_NAME) || true
	docker-compose up -d elastic

mq-reset:
	docker-compose rm -s -v -f mq || true
	docker volume rm $$(docker volume ls -q | grep mqdata | grep $$COMPOSE_PROJECT_NAME) || true
	docker-compose up -d mq

test:
	go test ./...

lint:
	golangci-lint run

test-api:
	# to install newman:
	# npm install -g newman
	newman run ./scripts/newman/tests.json -e ./scripts/newman/env.json

docs:
	# wget https://github.com/swaggo/swag/releases/download/v1.7.0/swag_1.7.0_Linux_x86_64.tar.gz
	# tar -zxvf swag_1.7.0_Linux_x86_64.tar.gz
	# sudo cp swag /usr/bin/swag
	cd cmd/api && swag init --parseDependency --parseInternal --parseDepth 2

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
	docker-compose down
	STABLE_TAG=$$(cat version.json | grep version | awk -F\" '{ print $$4 }' |  cut -d '.' -f1-2)
	TAG=$$STABLE_TAG $(MAKE) mq-reset
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

mq-list:
	docker-compose exec mq rabbitmqctl list_queues

ps:
	docker ps --format "table {{.Names}}\t{{.RunningFor}}\t{{.Status}}\t{{.Ports}}"

sandbox-images:
	docker-compose -f docker-compose.sandbox.yml pull

sandbox:
	COMPOSE_PROJECT_NAME=bcdbox docker-compose -f docker-compose.sandbox.yml up -d elastic mq db api indexer metrics gui

flextesa-sandbox:
	COMPOSE_PROJECT_NAME=bcdbox docker-compose -f docker-compose.sandbox.yml up -d

sandbox-down:
	COMPOSE_PROJECT_NAME=bcdbox docker-compose -f docker-compose.sandbox.yml down

sandbox-clear:
	COMPOSE_PROJECT_NAME=bcdbox docker-compose -f docker-compose.sandbox.yml down -v

gateway-images:
	docker-compose -f docker-compose.gateway.yml pull

gateway:
	COMPOSE_PROJECT_NAME=bcdhub docker-compose -f docker-compose.gateway.yml up -d

gateway-down:
	COMPOSE_PROJECT_NAME=bcdhub docker-compose -f docker-compose.gateway.yml down

gateway-clear:
	COMPOSE_PROJECT_NAME=bcdhub docker-compose -f docker-compose.gateway.yml down -v