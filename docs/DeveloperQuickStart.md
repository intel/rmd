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

RMD is published as open source project hosted on GitHub. Full source code can be downloaded using Go tools:


```bash
go get github.com/intel/rmd
```

or with git command:

```bash
git clone https://github.com/intel/rmd
```

It is also possible to download zipped sources from [RMD repository](https://github.com/intel/rmd) using GitHub web interface.

## Build & install rmd

Since release *v0.2* RMD is developed as a Go module. It means it can be downloaded into and build inside any system directory - not only in *$GOPATH/src/\<server\>/\<repository\>/* as it was before Go 1.11.

RMD supports only Linux OS and cannot be built for other operating systems. When building RMD using other system (cross compilation) please set *GOOS* environment variable to *linux*.

Assuming that RMD source code has been downloaded into *$HOME/sources/rmd*, build and installation process is as follow:

```bash
# Enter source directory
cd $HOME/sources/rmd

# Build RMD
make
# output binaries (rmd and gen_conf) can be found under $HOME/sources/rmd/build

# Install RMD
sudo make install

# RMD will be installed into /usr/local/sbin/ 
# Default configuration files will be placed under /etc/rmd
```

## Run rmd

To launch RMD in normal mode run:

```bash
sudo /usr/local/sbin/rmd
```

For testing purposes RMD can be launched in *debug mode* using command line param:

```bash
sudo /usr/local/sbin/rmd --debug
# or
sudo /usr/local/sbin/rmd -d
```

Please note that this mode does not provide REST API security (access control, connection encryption) and should not be used in production environment.

For more information about possible command line arguments please run:

```
$ sudo /usr/local/sbin/rmd --help
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

## Swagger

The API definitions are located under [swagger yaml](api/v1/swagger.yaml)

Upload [swagger yaml](api/v1/swagger.yaml) to http://editor.swagger.io/#!/ to generate
a client.
