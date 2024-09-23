#!/bin/bash

protoc --go_out=. --go_opt=paths=source_relative \
	--go-grpc_out=. --go-grpc_opt=paths=source_relative \
	protos/*.proto

python3 -m grpc_tools.protoc -I./protos \
	--python_out=./server/rpc/ --pyi_out=./server/rpc/ --grpc_python_out=./server/rpc/ \
	./protos/*.proto
