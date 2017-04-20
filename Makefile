dist: 
	./util/make-dist.sh
serve:
	goimports -w server/*.go
	go build -o serve server/*.go
all: serve dist
.PHONY: dist serve
