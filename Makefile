.PHONY: install

VERSION=`git describe --tags`

termage:
	go build -o termage -ldflags "-X main.version=${VERSION}" main.go

install: termage
	sudo cp termage /usr/local/bin
