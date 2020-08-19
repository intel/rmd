#Copyright 2017 Intel Corporation
#
#Licensed under the Apache License, Version 2.0 (the "License");
#you may not use this file except in compliance with the License.
#You may obtain a copy of the License at
#
#    http://www.apache.org/licenses/LICENSE-2.0
#
#Unless required by applicable law or agreed to in writing, software
#distributed under the License is distributed on an "AS IS" BASIS,
#WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
#See the License for the specific language governing permissions and
#limitations under the License.

# Start from a Debian image with the latest version of Go installed
# and a workspace (GOPATH) configured at /go.
FROM golang:1.13  
# Pull intel-cmt-cat
RUN mkdir -p /home/intel-cmt-cat \
           cd /home/intel-cmt-cat
RUN git clone https://github.com/intel/intel-cmt-cat.git
WORKDIR /go/intel-cmt-cat
RUN make install
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
