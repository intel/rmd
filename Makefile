OS = $(shell uname -s | tr '[:upper:]' '[:lower:]')
ARCH = $(shell uname -m)
.PHONY: deps check build test-unit test-func install clean

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
install: build
	mkdir -p /usr/local/sbin
	cp build/$(OS)/$(ARCH)/rmd /usr/local/sbin/
	bash -c "./scripts/install.sh --skip-pam-userdb"
clean:
	rm -rf build
