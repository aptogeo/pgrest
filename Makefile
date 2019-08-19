all:
	go get
	go test
	go test -short -race
