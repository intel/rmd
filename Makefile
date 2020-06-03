OS = $(shell uname -s | tr '[:upper:]' '[:lower:]')
ARCH = $(shell uname -m)
VERSION = $(shell cat VERSION | sed -e 's/^v//')
OUTPUT_DIR=build/$(OS)/$(ARCH)
RMD_DIR = rmd-$(VERSION)
go-bin-deb = $(GOPATH)/bin/go-bin-deb
go-bin-rpm = $(GOPATH)/bin/go-bin-rpm

BUILD_TYPE ?= standard

.PHONY: build check clean deb deps docker install rpm test-func test-unit pstate pstateinstall

default: deps build

deps:
	bash -c "./scripts/deps.sh"
check:
	bash -c "./scripts/hacking_v2.sh"
test-unit: deps
	bash -c "./scripts/test.sh -u -b $(BUILD_TYPE)"
test-func: deps build
	bash -c "./scripts/test.sh -f"
build: deps
	bash -c "./scripts/build.sh -b $(BUILD_TYPE)"
deb: build
ifeq (x86_64, $(ARCH))
	@$(go-bin-deb) generate --arch amd64 --version $(PACKAGE_VERSION)
else
	@$(go-bin-deb) generate --arch $(ARCH) --version $(PACKAGE_VERSION)
endif
rpm: build
	@$(go-bin-rpm) generate --arch $(ARCH) --version $(PACKAGE_VERSION) --output rmd-$(PACKAGE_VERSION).$(ARCH).rpm
install: build
	mkdir -p $(DESTDIR)/usr/bin
	cp $(OUTPUT_DIR)/rmd $(DESTDIR)/usr/bin/rmd
	cp $(OUTPUT_DIR)/gen_conf $(DESTDIR)/usr/bin/gen_conf
	bash -c "./scripts/install.sh --skip-pam-userdb"
package:
	mkdir -p RMD_DIR
	rsync -avr --exclude=$(RMD_DIR),.git,build * $(RMD_DIR)
	tar -zcvf $(RMD_DIR).tar.gz $(RMD_DIR)/
	rm -rf $(RMD_DIR)/
	rpmdev-setuptree
	cp packaging/rmd.spec $(HOME)/rpmbuild/SPECS
	mv $(RMD_DIR).tar.gz $(HOME)/rpmbuild/SOURCES
	bash -c "./packaging/download_extra_packages.sh $(PWD)/packages/ $(HOME)/rpmbuild/SOURCES"
	rpmbuild -ba $(HOME)/rpmbuild/SPECS/rmd.spec
docker:
	@docker build -t rmd .
clean:
	rm -rf build pkg-build

