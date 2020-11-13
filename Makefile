include .env
export $(shell sed 's/=.*//' .env)

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
	docker exec -it $$BCD_ENV-api nginx
endif
	#docker restart $$BCD_ENV-gui

migration:
ifeq ($(BCD_ENV), development)
	cd scripts/migration && go run .
else
	docker exec -it $$BCD_ENV-api migration
endif

rollback:
	cd scripts/esctl && go run . rollback -n $(NETWORK) -l $(LEVEL)

remove:
	cd scripts/esctl && go run . remove -n $(NETWORK) 

s3-creds:
	docker exec -it $$BCD_ENV-elastic bash -c 'bin/elasticsearch-keystore add --force --stdin s3.client.default.access_key <<< "$$AWS_ACCESS_KEY_ID"'
	docker exec -it $$BCD_ENV-elastic bash -c 'bin/elasticsearch-keystore add --force --stdin s3.client.default.secret_key <<< "$$AWS_SECRET_ACCESS_KEY"'
ifeq ($(BCD_ENV), development)
	cd scripts/esctl && go run . reload_secure_settings
else
	docker exec -it $$BCD_ENV-api esctl reload_secure_settings
endif

s3-repo:
ifeq ($(BCD_ENV), development)
	cd scripts/esctl && go run . create_repository
else
	docker exec -it $$BCD_ENV-api esctl create_repository
endif

s3-restore:
ifeq ($(BCD_ENV), development)
	cd scripts/esctl && go run . restore
else
	docker exec -it $$BCD_ENV-api esctl restore
endif

s3-snapshot:
ifeq ($(BCD_ENV), development)
	cd scripts/esctl && go run . snapshot
else
	docker exec -it $$BCD_ENV-api esctl snapshot
endif

s3-policy:
ifeq ($(BCD_ENV), development)
	cd scripts/esctl && go run . set_policy
else
	docker exec -it $$BCD_ENV-api esctl set_policy
endif

es-reset:
	docker stop $$BCD_ENV-elastic || true
	docker rm $$BCD_ENV-elastic || true
	docker volume rm $$COMPOSE_PROJECT_NAME_esdata || true
	docker-compose up -d elastic

clearmq:
	docker exec -it $$BCD_ENV-mq rabbitmqctl stop_app
	docker exec -it $$BCD_ENV-mq rabbitmqctl reset
	docker exec -it $$BCD_ENV-mq rabbitmqctl start_app

test:
	go test ./...

lint:
	golangci-lint run
  
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

latest:
	docker-compose up -d

upgrade:
	$(MAKE) clearmq
	docker-compose down
	TAG=$$STABLE_TAG $(MAKE) es-reset
	docker-compose up -d db mq

restart:
	docker-compose restart api metrics indexer compiler

release:
	BCDHUB_VERSION=$$(cat version.json | grep version | awk -F\" '{ print $$4 }')
	git tag $$BCDHUB_VERSION && git push origin $$BCDHUB_VERSION

db-dump:
	docker exec -it $$BCD_ENV-db pg_dump -c bcd > dump_`date +%d-%m-%Y"_"%H_%M_%S`.sql

db-restore:
	docker exec -i $$BCD_ENV-db psql --username $$POSTGRES_USER -v ON_ERROR_STOP=on bcd < $(BACKUP)

ps:
	docker ps --format "table {{.Names}}\t{{.RunningFor}}\t{{.Status}}\t{{.Ports}}"

sandbox:
	COMPOSE_PROJECT_NAME=bcd-box docker-compose -f docker-compose.sandbox.yml up -d --build

sandbox-down:
	COMPOSE_PROJECT_NAME=bcd-box docker-compose -f docker-compose.sandbox.yml down
