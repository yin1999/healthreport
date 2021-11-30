# make
all: dep build

dep:
	@echo Downloading dependencies...
	@go mod download

build:
	@echo Building...
	@go run _script/make.go -goos="${TARGET}" -version="${VERSION}" -goarch="${ARCH}" -goarm="${ARM}"
	@echo Done.
