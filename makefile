gen/todo:
	protoc -Iproto --go_out=proto \
		--go_opt=paths=source_relative \
		--go-grpc_out=proto \
		--go-grpc_opt=paths=source_relative \
		--validate_out="lang=go,paths=source_relative:proto" \
		proto/todo/v1/*.proto

run/server:
	go run ./server 0.0.0.0:50051 0.0.0.0:50052

run/client:
	go run ./client 0.0.0.0:50051

kub/apply:
	kubectl apply -f k8s/server.yml
	kubectl apply -f k8s/client.yml

kub/delete:
	kubectl delete -f k8s/server.yml
	kubectl delete -f k8s/client.yml
