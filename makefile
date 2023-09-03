gen/todo:
	protoc --go_out=. \
		--go_opt=paths=source_relative \
		--go-grpc_out=. \
		--go-grpc_opt=paths=source_relative \
		proto/todo/v1/*.proto

run/server:
	go run ./server 0.0.0.0:50051

run/client:
	go run ./client 0.0.0.0:50051

kind/create:
	kind create cluster --config k8s/kind.yaml

kind/delete:
	kind delete cluster