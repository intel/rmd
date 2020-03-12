#!/usr/bin/env bash

GO_MINOR_VERSION=$(go version | cut -c 14- | cut -d' ' -f1 | cut -d'.' -f2)

if [ ${GO_MINOR_VERSION} -ge 11 ]; then
	export GO111MODULE=on
else

	BASE=$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)
	. $BASE/go-env

	go get github.com/Masterminds/glide && glide install

	# Install go-bin-deb
	# mkdir -p $GOPATH/src/github.com/mh-cbon/go-bin-deb
	# cd $GOPATH/src/github.com/mh-cbon/go-bin-deb
	# git clone https://github.com/mh-cbon/go-bin-deb.git .
	# glide install
	# go install
	go get github.com/mh-cbon/go-bin-deb
	go get github.com/mh-cbon/go-bin-rpm
fi
