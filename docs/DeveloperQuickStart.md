# RMD Developer Quickstart Guide

## Prepare GO development environment

Follow https://golang.org/doc/install to install golang.
Make sure you have your $GOPATH, $PATH setup correctly

e.g.
```
export GOPATH=$HOME/go
export PATH=$PATH:/usr/local/go/bin:$GOPATH/bin
# check if your go environment variables are set correctly
go env
```

## Get rmd

```
$ go get github.com/intel/rmd
```

## Build & install rmd

```
# build rmd
$ cd ${GOPATH}/src/github.com/intel/rmd
$ make
# You will fine the binary under ${GOPATH}/src/github.com/intel/rmd/build

# install RMD
$ sudo make install
# this will install RMD into /usr/local/sbin/ along with some default
configuration file under /etc/rmd
```

## Run rmd

```
$ sudo /usr/local/sbin/rmd --help
$ sudo /usr/local/sbin/rmd
```

## Commit code

Bash shell script `hacking.sh` checks coding style using `go fmt` and `golint`.
Before you commit your changes, run `hacking.sh` in scripts directory,
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
