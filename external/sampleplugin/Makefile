# This is Makefile for building RMD plugin as loadable .so file
BUILD_DIR ?= ./build
PLUGIN_SO_NAME ?= sampleplugin.so
INSTALL_PATH ?= /etc/rmd/plugins

.PHONY: all clean $(BUILD_DIR)/$(PLUGIN_SO_NAME) install

all: $(BUILD_DIR)/$(PLUGIN_SO_NAME)

clean:
	@rm -rf $(BUILD_DIR)

$(BUILD_DIR)/$(PLUGIN_SO_NAME): $(BUILD_DIR)
	go build -buildmode=plugin -o $(BUILD_DIR)/$(PLUGIN_SO_NAME) ./

$(BUILD_DIR):
	@mkdir -p $(BUILD_DIR)

install:
	@mkdir -p $(INSTALL_PATH)
	cp $(BUILD_DIR)/$(PLUGIN_SO_NAME) $(INSTALL_PATH)/

test:
	go test -v ./
