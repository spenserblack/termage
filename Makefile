PACKAGE="github.com/spenserblack/termage"
VERSION=`git describe --tags`

termage:
	go build -o termage -ldflags "-X ${PACKAGE}/cmd.Version=${VERSION}" main.go
