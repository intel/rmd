# RMD Developer Quickstart Guide

## Prepare GO development environment

Follow https://golang.org/doc/install to install golang.
Make sure you have your $GOPATH, $PATH setup correctly

*Note: only support build RMD on linux host(GOOS=Linux)*

e.g.
```
export GOPATH=$HOME/go
export PATH=$PATH:/usr/local/go/bin:$GOPATH/bin
# check if your go environment variables are set correctly
go env
```

As RMD is prepared as "Go module", it needs at least Go 1.11 version to be build properly. However, it is highly recommended to use version not lower than 1.13 due to multiple bugs fixed in the standard library.

## Get rmd

```
$ go get github.com/intel/rmd
```

## Build & install rmd

```
# build rmd, only support linux host(GOOS=Linux)
$ cd ${GOPATH}/src/github.com/intel/rmd
$ make
# You will find the rmd binary under ${GOPATH}/src/github.com/intel/rmd/build

# install RMD
$ sudo make install
# This will install RMD into /usr/local/sbin/ along with some default
configuration files under /usr/local/etc/rmd
```

## Run rmd

```
$ sudo /usr/local/sbin/rmd --help
$ sudo /usr/local/sbin/rmd
```

## Commit code

Bash shell script `hacking_v2.sh` checks coding style using `go fmt` and `golint`.
Before you commit your changes, run `hacking_v2.sh` in scripts directory,
and address errors before you push your changes.

Alternatively, you can run `make check` to do a full static code checking.

## Test

Bash shell script `test.sh` is a helper script to do unit testing and
functional testing.

`sudo -E scripts/test.sh -u` to run all unit test cases or just do
`sudo make test-unit`

P.S. There is an issue when do the unit test case, some of the cases depend
on configuration file, to pass unit test case, better to install rmd first
by `sudo make install` or manually copy etc/rmd to /etc/rmd.

To run functional testing, you need to install ginkgo by:

```
$ go get github.com/onsi/ginkgo/ginkgo
```

Run `sudo make test-func` to run functional test, alternatively using the
following schell scripts command line.

`sudo -E ./scripts/test.sh -i` to run all functional test cases.
`sudo -E ./scripts/test.sh -i -s` to run all functional test cases with certificate
based https support.
`sudo -E ./scripts/test.sh -i -s -nocert` to run all functional test cases with PAM
based https support.

Read test.sh to understand what functional test cases do.

## Glide

Use glide (https://github.com/Masterminds/glide) to manage dependencies.

## Swagger

The API definitions are located under [swagger yaml](api/v1/swagger.yaml)

Upload [swagger yaml](api/v1/swagger.yaml) to http://editor.swagger.io/#!/ to generate
a client.
