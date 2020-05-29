include .env
export $(shell sed 's/=.*//' .env)

api:
	cd cmd/api && go run . -f config.yml -f config.dev.yml

indexer:
	cd cmd/indexer && go run . -f config.yml -f config.dev.yml

metrics:
	cd cmd/metrics && go run . -f config.yml -f config.dev.yml

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

deploy: export TAG=$(shell git pull -q && git describe --abbrev=0 --tags)
deploy:
	git pull
	docker-compose pull
	docker-compose up -d
	docker-compose ps

task:
	cd scripts/ml && go run . -f ../config.yml
  
docs:
	# wget https://github.com/swaggo/swag/releases/download/v1.6.6/swag_1.6.6_Linux_x86_64.tar.gz
	# tar -zxvf swag_1.6.6_Linux_x86_64.tar.gz
	# sudo cp swag /usr/bin/swag
	cd cmd/api && swag init --parseDependency

images:
	docker-compose build