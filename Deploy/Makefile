TAG ?= 0.0.1

compile:
	protoc --go_out=. --go_opt=paths=source_relative \
		--go-grpc_out=. --go-grpc_opt=paths=source_relative \
		api/v1/health.proto

build-docker:
	docker build -t github.com/gergeshany/event-streaming-system:$(TAG) -f Dockerfile ..