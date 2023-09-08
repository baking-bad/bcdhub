-include .env
export $(shell sed 's/=.*//' .env)

LATEST_DUMP=/tmp/dump_latest.gz
BACKUP?=dump_latest.gz

.ONESHELL:

api:
	docker-compose up -d db
	cd cmd/api && go run .

indexer:
	docker-compose up -d db
	cd cmd/indexer && go run .

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

s3-db-restore:
	echo "Database restore..."
ifeq (,$(wildcard $(LATEST_DUMP)))
	aws s3 cp --profile bcd s3://bcd-db-snaps/$(BACKUP) $(LATEST_DUMP)
endif

	docker-compose exec -T db dropdb -U $(POSTGRES_USER) --if-exists $(POSTGRES_DB)
	gunzip -dc $(LATEST_DUMP) | docker-compose exec -T db psql -U $(POSTGRES_USER) -v ON_ERROR_STOP=on $(POSTGRES_DB)
	rm $(LATEST_DUMP)

s3-db-snapshot:
	echo "Database snapshot..."
	docker-compose exec db pg_dump $(POSTGRES_DB) --create -U $(POSTGRES_USER) | gzip -c > $(LATEST_DUMP)	
	aws s3 mv --profile bcd $(LATEST_DUMP) s3://bcd-db-snaps/dump_latest.gz

s3-list:
	echo "Database snapshots"
	aws s3 ls --profile bcd s3://bcd-db-snaps

test:
	go test ./...

lint:
	golangci-lint run

test-api:
	# to install newman:
	# npm install -g newman
	newman run ./scripts/newman/tests.json -e ./scripts/newman/env.json

stable:
	TAG=master docker-compose up -d api indexer

db-dump:
	docker-compose exec db pg_dump -c $(POSTGRES_DB) > dump_`date +%d-%m-%Y"_"%H_%M_%S`.sql

db-restore:
	docker-compose exec -T db psql --username $(POSTGRES_USER) -v ON_ERROR_STOP=on $(POSTGRES_DB) < $(BACKUP)

ps:
	docker ps --format "table {{.Names}}\t{{.RunningFor}}\t{{.Status}}\t{{.Ports}}"

sandbox-pull:
	TAG=4.4.0 docker-compose -f docker-compose.flextesa.yml pull

flextesa-sandbox:
	COMPOSE_PROJECT_NAME=bcdbox TAG=4.4.0 docker-compose -f docker-compose.flextesa.yml up -d

sandbox-down:
	COMPOSE_PROJECT_NAME=bcdbox docker-compose -f docker-compose.flextesa.yml down

sandbox-clear:
	COMPOSE_PROJECT_NAME=bcdbox docker-compose -f docker-compose.flextesa.yml down -v

generate:
	go generate ./...