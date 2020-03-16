include .env
export $(shell sed 's/=.*//' .env)

up:
	docker-compose up -d

build:
	docker-compose build

api:
	cd cmd/api && go run .
