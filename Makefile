dist: 
	./util/make-dist.sh
server:
	go build -o serve server/*.go
all: server dist
.PHONY: dist server
