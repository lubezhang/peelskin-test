run:
	clear
	@go run main.go

build: clean build_osx build_linux build_windows

build_osx:
	env GOOS=darwin GOARCH=amd64 go build -o ./bin/hlsdl_osx ./main.go
	@md5 ./bin/hlsdl_osx

build_linux:
	env GOOS=linux GOARCH=amd64 go build -o ./bin/hlsdl_linux ./main.go
	@md5 ./bin/hlsdl_linux

build_windows:
	env GOOS=windows GOARCH=amd64 go build -o ./bin/hlsdl_windows.exe ./main.go
	@md5 ./bin/hlsdl_windows.exe

clean:
	@rm -rf bin/
	@go clean

test:
	clear
	@go test ./...
	@#go test -v ./tests/...
