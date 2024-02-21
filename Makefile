PROJECT_NAME = "cproject"
MAKEFILE_DIR = $(shell dirname $(realpath $(firstword $(MAKEFILE_LIST))))
DIST_PATH = ./dist
ARTIFACT_PATH = $(DIST_PATH)/$(PROJECT_NAME)
BIN_PATH = ./bin

GO_MOD_PACKAGE = "github.com/marklap/cproject"
GO_FILES = "./cmd/"
GO_BUILD_FLAGS=-installsuffix 'static' -ldflags "-s -w"

.DEFAULT_GOAL := build

.PHONY: go-mod-init
go-mod-init::
ifeq (,$(wildcard ./go.mod))
	go mod init $(GO_MOD_PACKAGE)
endif

.PHONY: go-mod-update
go-mod-update:: go-mod-init
	go mod tidy

.PHONY: build
build:: go-mod-update
	GOOS=linux CGO_ENABLED=0 go build -o $(ARTIFACT_PATH).linux_amd64 $(GO_BUILD_FLAGS) $(GO_FILES)
	GOOS=darwin CGO_ENABLED=0 go build -o $(ARTIFACT_PATH).darwin_amd64 $(GO_BUILD_FLAGS) $(GO_FILES)
	GOOS=windows CGO_ENABLED=0 go build -o $(ARTIFACT_PATH).windows_amd64.exe $(GO_BUILD_FLAGS) $(GO_FILES)

.PHONY: dist-zip
dist-zip::
	rm -f $(ARTIFACT_PATH).zip
	zip -rj $(ARTIFACT_PATH).zip $(DIST_PATH)

.PHONY: dist
dist:: build dist-zip

.PHONY: install
install:: build
	@mkdir -p $(BIN_PATH)
	rm -Rf $(BIN_PATH)/$(PROJECT_NAME)
	cp $(ARTIFACT_PATH).linux_amd64 $(BIN_PATH)/$(PROJECT_NAME)
	chmod u+x $(BIN_PATH)/$(PROJECT_NAME)

.PHONY: clean
clean::
	rm -Rf $(DIST_PATH)