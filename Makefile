# make
all: dep build

dep:
	@echo Downloading dependencies...
	@go mod download

build:
	@echo Building...
	@go run _script/make.go -GOOS="${TARGET}" -version="${VERSION}"
	@echo Done.
