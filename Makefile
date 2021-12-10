
.PHONY: build

build:
	go build -trimpath -ldflags="-s -w" -o ./bin/unneko ./cmd
	GOOS=darwin GOARCH=amd64 go build -trimpath -ldflags="-s -w" -o ./bin/unneko-darwin-x64 ./cmd
	GOOS=darwin GOARCH=arm64 go build -trimpath -ldflags="-s -w" -o ./bin/unneko-darwin-arm64 ./cmd
	GOOS=linux GOARCH=amd64 go build -trimpath -ldflags="-s -w" -o ./bin/unneko-linux-x64 ./cmd
	GOOS=windows GOARCH=amd64 go build -trimpath -ldflags="-s -w" -o ./bin/unneko-win-x64.exe ./cmd
