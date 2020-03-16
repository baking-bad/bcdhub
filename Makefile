include .env
export $(shell sed 's/=.*//' .env)

up:
	docker-compose up -d

build:
	docker-compose build

deploy:
	docker-compose pull
	docker-compose up -d

api:
	cd cmd/api && go run .
