include .env
export $(shell sed 's/=.*//' .env)

deploy: export TAG=$(shell git pull -q && git describe --abbrev=0 --tags)
deploy:
	git pull
	docker-compose pull
	docker-compose -f docker-compose.yml -f docker-compose.prod.yml up -d
	docker ps

api:
	cd cmd/api && go run . -f config.yml -f config.dev.yml

indexer:
	cd cmd/indexer && go run . -f config.yml -f config.dev.yml

metrics:
	cd cmd/metrics && go run . -f config.yml -f config.dev.yml

clearmq:
	docker exec -it bcd-mq rabbitmqctl stop_app
	docker exec -it bcd-mq rabbitmqctl reset
	docker exec -it bcd-mq rabbitmqctl start_app

aliases:
	cd scripts/aliases && go run . -f ../config.yml

rollback:
	cd scripts/rollback && go run . -f ../config.yml -n $(NETWORK) -l $(LEVEL)

migration:
	cd scripts/migration && go run . -f ../config.yml

upd:
	docker-compose -f docker-compose.yml docker-compose.prod.yml up -d --build

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