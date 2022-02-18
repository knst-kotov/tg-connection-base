redis:
	docker run --name redis -p 6379:6379 -d redis

run:
	go run main.go -test

up:
	docker-compose up --build

.PHONY: redis run up