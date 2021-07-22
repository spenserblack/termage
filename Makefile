.PHONY: install

VERSION=`git describe --tags`

termage:
	go build -ldflags "-X main.version=${VERSION}" termage.go

install: termage
	sudo cp termage /usr/local/bin
