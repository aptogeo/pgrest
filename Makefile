all:
	go get
	go test
	go test transactional/**
	go test -short -race
