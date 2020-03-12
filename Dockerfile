# Start from a Debian image with the latest version of Go installed
# and a workspace (GOPATH) configured at /go.
FROM golang:1.13  

# Copy the local package files to the container's workspace.
WORKDIR /go/src/github.com/intel/rmd
COPY . .

#Add proxy settings below if behind a proxy.
#ENV http_proxy=
#ENV https_proxy=
#ENV ftp_proxy=
#ENV socks_proxy=
#ENV no_proxy=

RUN apt update && apt install openssl libpam0g-dev db-util -y && \
        rm -rf /var/lib/apt/lists/*

RUN make install && make clean

#RUN mount -t resctrl resctrl /sys/fs/resctrl
# what etc should we use?
# log

# Run the outyet command by default when the container starts.
ENTRYPOINT ["/usr/bin/rmd","-d","--address","0.0.0.0"]

# Document that the service listens on port 8080.
#EXPOSE 8081
