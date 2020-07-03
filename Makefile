setup:
	@echo -e "\033[33mPREREQUISITES : git sed go php composer ruby node npm\033[0m"
	@echo -e "\033[33mPREREQUISITES : sudo npm i -g bower gulp csso polymer-bundler@2.0.0-pre12\033[0m"
	go install github.com/mattn/go-sqlite3
	go install github.com/dayvonjersen/vibrant
	npm i
	bower i
	gulp data
	cd util && composer install

dist: 
	./util/make-dist.sh
serve:
	goimports -w server/*.go
	go build -o serve server/*.go
all: serve dist
.PHONY: dist serve
