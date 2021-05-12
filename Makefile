all:
	go get
	go get github.com/stretchr/testify
	go get github.com/google/uuid
	go test
	go test transactional/**
	go test -short -race
