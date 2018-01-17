OS = $(shell uname -s | tr '[:upper:]' '[:lower:]')
ARCH = $(shell uname -m)
VERSION = $(shell git describe --abbrev=0 --tags | grep -Eo '[0-9]{1,3}\.[0-9]{1,3}')
PACKAGE_VERSION = $(VERSION)".0"

go-bin-deb = $(GOPATH)/bin/go-bin-deb
go-bin-rpm = $(GOPATH)/bin/go-bin-rpm

.PHONY: build check clean deb deps docker install rpm test-func test-unit

default: deps build

deps:
	bash -c "./scripts/deps.sh"
check:
	bash -c "./scripts/hacking.sh -f"
test-unit: deps
	bash -c "./scripts/test.sh -u"
test-func: deps
	bash -c "./scripts/test.sh -f"
build: deps
	bash -c "./scripts/build.sh"
deb: build
ifeq (x86_64, $(ARCH))
	@$(go-bin-deb) generate --arch amd64 --version $(PACKAGE_VERSION)
else
	@$(go-bin-deb) generate --arch $(ARCH) --version $(PACKAGE_VERSION)
endif
rpm: build
	@$(go-bin-rpm) generate --arch $(ARCH) --version $(PACKAGE_VERSION) --output rmd-$(PACKAGE_VERSION).$(ARCH).rpm
install: build
	mkdir -p /usr/local/sbin
	cp build/$(OS)/$(ARCH)/rmd /usr/local/sbin/
	bash -c "./scripts/install.sh --skip-pam-userdb"
docker:
	@docker build -t rmd .
clean:
	rm -rf build pkg-build
