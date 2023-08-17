.PHONY: build

build:
	go build

install: build
	mkdir -p ~/bin/
	cp aws-fail-az ~/bin/aws-fail-az
	chmod 700 ~/bin/aws-fail-az
