include .env
export $(shell sed 's/=.*//' .env)

api:
	cd cmd/api && go run . -f config.yml -f config.dev.yml

indexer:
	cd cmd/indexer && go run . -f config.yml -f config.dev.yml

metrics:
	cd cmd/metrics && go run . -f config.yml -f config.dev.yml

compiler:
	cd cmd/compiler && go run . -f config.yml -f config.dev.yml

aliases:
	cd scripts/aliases && go run . -f ../config.yml

sitemap:
	cd scripts/sitemap && go run . -f ../config.yml

rollback:
	cd scripts/rollback && go run . -f ../config.yml

migration:
	cd scripts/migration && go run . -f ../config.yml

es-aws:
	cd scripts/es-aws && go build .

s3-creds: es-aws
	docker exec -it bcd-elastic bash -c 'bin/elasticsearch-keystore add --stdin s3.client.default.access_key <<< "$$AWS_ACCESS_KEY_ID"'
	docker exec -it bcd-elastic bash -c 'bin/elasticsearch-keystore add --stdin s3.client.default.secret_key <<< "$$AWS_SECRET_ACCESS_KEY"'
	./scripts/es-aws/es-aws -a reload_secure_settings -f scripts/config.yml

s3-repo: es-aws
	./scripts/es-aws/es-aws -a create_repository -f scripts/config.yml

s3-restore: es-aws
	./scripts/es-aws/es-aws -a delete_indices -f scripts/config.yml
	./scripts/es-aws/es-aws -a restore -f scripts/config.yml

s3-snapshot: es-aws
	./scripts/es-aws/es-aws -a snapshot -f scripts/config.yml

s3-policy: es-aws
	./scripts/es-aws/es-aws -a set_policy -f scripts/config.yml

latest:
	git tag latest -f && git push origin latest -f

es-reset:
	docker stop bcd-elastic || true
	docker rm bcd-elastic || true
	docker volume rm bcdhub_esdata || true
	docker-compose up -d elastic

clearmq:
	docker exec -it bcd-mq rabbitmqctl stop_app
	docker exec -it bcd-mq rabbitmqctl reset
	docker exec -it bcd-mq rabbitmqctl start_app

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
	docker-compose down
	TAG=$$STABLE_TAG $(MAKE) es-reset

restart:
	docker-compose restart bcd-api bcd-metrics bcd-indexer

release:
	BCDHUB_VERSION=$$(cat version.json | grep version | awk -F\" '{ print $$4 }')
	git tag $$BCDHUB_VERSION && git push origin $$BCDHUB_VERSION

db-dump:
	docker exec -it bcd-db pg_dump -c bcd > dump_`date +%d-%m-%Y"_"%H_%M_%S`.sql

db-restore:
	docker exec -i bcd-db psql --username $$POSTGRES_USER -v ON_ERROR_STOP=on bcd < $(BACKUP)
