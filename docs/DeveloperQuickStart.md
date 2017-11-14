# RMD Developer Quickstart Guide

## Prepare GO development environment

Follow https://golang.org/doc/install to install golang.
Make sure you have your $GOPATH, $PATH setup correctly

```
e.g.
export GOPATH=$HOME/go
export PATH=$PATH:/usr/local/go/bin:$GOPATH/bin
# check if your go environment variables are set correctly
go env
```

## Clone rmd code

Clone or copy the code into $GOPATH/src/github.com/intel/rmd

## Build & install rmd

```
$ go get github.com/Masterminds/glide && glide install

# generage configuration file
$ go run cmd/gen_conf.go

# install RMD into $GOPATH/bin
$ ./install-deps.sh
# To skip setting up PAM Berkeley DB users supply.
$ ./install-deps.sh --skip-pam-userdb
```

## Run rmd

```
$ $GOPATH/bin/rmd --help
$ $GOPATH/bin/rmd
```

## Commit code

Bash shell script `hacking.sh` checks coding style using `go fmt` and `golint`.

Before you commit your changes, run `./hacking.sh` and address errors before you push your changes.

## Test

Bash shell script `test.sh` is a helper script to do unit testing and functional testing.

`./test.sh -u` to run all unit test cases.
`./test.sh -i` to run all functional test cases.
`./test.sh -i -s` to run all functional test cases with certificate based https support.
`./test.sh -i -s -nocert` to run all functional test cases with PAM based https support.

Read test.sh to understand what functional test cases do.

## Glide

Use glide (https://github.com/Masterminds/glide) to manage dependencies.

## Swagger

The API definitions are located under docs/v1/swagger.yaml

Upload docs/api/v1/swagger.yaml to http://editor.swagger.io/#!/ to generate a client.
