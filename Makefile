SOURCE_CODE := $(wildcard cmd/*.go) $(wildcard internal/*.go)

GOOS := $(shell go env GOOS)
GOARCH := $(shell go env GOARCH)

BINARY_NAME := "aish-$(GOOS)-$(GOARCH)"

ifeq ($(GOOS),windows)
BINARY_NAME := "$(BINARY_NAME).exe"
endif

build/$(BINARY_NAME): $(SOURCE_CODE)
	@CGO_ENABLED=0 GOOS=$(GOOS) GOARCH=$(GOARCH) go build -o ./build/$(BINARY_NAME) ./cmd/aish/main.go

.PHONY: install

install: build/$(BINARY_NAME)
ifeq ($(GOOS),windows)
	@if not exist "C:\Windows\System32" mkdir "C:\Windows\System32"
	@copy /Y .\build\$(BINARY_NAME) "C:\Windows\System32\aish.exe"
else
	@sudo rm -f /usr/local/bin/$(BINARY_NAME)
	@sudo cp ./build/$(BINARY_NAME) /usr/local/bin/aish
endif
