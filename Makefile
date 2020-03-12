include .env
export $(shell sed 's/=.*//' .env)

api:
	cd cmd/api && go run .