# RMD User Guide:

## Prerequisite

To use RMD, the users need to meet several requirements for both hardware and
software. Please read [this doc](../docs/Prerequisite.md) for more details.

## Installation

Currently, RMD supports only executable binary file downloading, and the users
need to install it to system path manually.

### Download

The RMD executable binary files are hosted in github.com infrastructure, and
you can find the download links from [release notes](https://github.com/intel/rmd/releases).

### Setup configuration file

RMD has some default configurations, you may also find some configure file
example from this [samples](../etc/rmd).

Most of the configuration files are in [TOML](https://github.com/toml-lang/toml) format.
Users can create their own configuration files by referring the following
sample. And RMD also provides the script(cmd/gen_conf.go) to generate it
by probing the local host platform capabilities for proper CachePool settings.

RMD will try to search `/etc/rmd` to find
configure files, put configuration files into these directory or RMD will
use default configurations.

Default location for main RMD configuration file is: /etc/rmd/rmd.toml

#### Cache module configuration and usage

Below extract from main configuration file (rmd.toml) presents sample Cache module configuration

```
[OSGroup] # mandatory
cacheways = 1
cpuset = "0-1"

[InfraGroup] # optional
cacheways = 19
cpuset = "2-3"
# arrary or comma-separated values
tasks = ["ovs*"] # just support Wildcards

[CachePool]
shrink = false # whether allow to shrink cache ways in best effort pool
max_allowed_shared = 10 # max allowed workload in shared pool, default is 10
guarantee = 10
besteffort = 7
shared = 2

```

Comments of the directives in the above conf:

* OSGroup: cache ways reserved for operating system usage.
* InfraGroup: infrastructure tasks group, user can specify task binary name
              the cache ways will be shared with other workloads
* CachePool: cache allocation will happened in the pools.
    - shrink: whether shrink cache ways which already allocated to workload in
            "besteffort" pool.
    - max_allowed_shared: max allowed workloads in shared cache pool
    - guarantee: allocate cache for workload max_cache == min_cache > 0
    - besteffort: allocate cache for workload max_cache > min_cache > 0
    - shared: allocate cache for workload max_cache == min_cache = 0

On a host which support max 20 cache ways, for this configuration file,
we will have follow cache bit mask layout:
```
OSGroup:    0000 0000 0000 0000 0001
InfraGroup: 1111 1111 1111 1111 1110
```

Available CBM in Cache Pools initially:
```
guarantee:  0000 0000 0111 1111 1110
besteffort: 0011 1111 1000 0000 0000
shared:     1100 0000 0000 0000 0000
```

#### P-State module configuration and usage

P-State module is designed to change CPU core frequency based on it's load defined as a branch prediction miss/hits ratio.
If current branch ratio is higher than selected threshold then CPU core frequency is set to predefined maximal value (can be CPU max frequency).
When branch ratio on given core falls below the selected threshold then core frequency is set to minimal.

Below extract from main configuration file (*rmd.toml*) presents sample configuration for optional P-State module.

```
[pstate]
## Simple plugin enable/disable flag
enabled = true
## path to loadable plugin file (.so library) with P-State implementation
path = "/etc/rmd/plugins/pstatelib.so"
## port number with plugin's http server (REST API)
port = 8080
```

This module is not an integral part of RMD binary but a loadable external plugin. To have P-State module working it is necessary to add section from above to *rmd.toml* and also provide *.so* file with plugin implementation.

While providing plugin binary and setting the ***path*** it is crucial to remember that this location (thus this file) has to be readable for both *root* and *rmd* user. Otherwise RMD processes will not be able to load and use plugin library.

If ***enabled*** flag is set to *false* plugin file existence will not be checked and plugin will not be loaded (even if exists).

***port*** is the number of listening port where branch_monitor's http server with REST API is available.

### Prepare the credentials

For security considerations, RMD daemon runs the RESTful API service as a
dedicated Unix user 'rmd' and performs privileged operations (PAM
authentication or resctrl) as Unix 'root' user. RMD users need to prepare
several things by themselves.

1. Create user and group 'rmd' in the target Linux system

    ```shell
    $ sudo useradd rmd
    ```
    P.S. RMD itself will create a rmd user if that user does not existed.

2. Ensure the following packages are installed on your target system for PAM

    Debian or Ubuntu:
    ```shell
    $ sudo apt-get install openssl libpam0g-dev db-util
    ```
    Redhat(s)
    ```shell
    $ sudo dnf install openssl pam-devel db4-utils
    ```
3. Besides host Unix credentials, RMD can also use dedicated credentials setup
   in a Berkeley database file.

    To run this script to setup users in Berkeley database:
    ```shell
    $ sudo /usr/share/rmd/setup_rmd_users.sh
    ```
    *Note: Only root user can setup or access users in Berkeley database*

## Run in docker container

### Build RMD docker image
Make sure you have install docker service.

```
$ sudo make docker
# Run RMD in docker
$ sudo docker run  --privileged -v /proc:/proc \
        -v /sys/fs/resctrl:/sys/fs/resctrl:rw \
        --address 0.0.0.0
```

## Run the service

Launch RMD manually, by specifying configuration directory:

```shell
$ sudo rmd --conf-dir /etc/rmd
```

RMD can be launched in debug mode that exposes RESTAPI service on HTTP.
By default it is launched on port 8081. (http://127.0.0.1:8081)

```shell
$ sudo rmd --debug
```

## RMD service usages

In this section, RMD is launched in debug mode and API calls uses insecure HTTP connection. See [possible connection types](#supported-rmd-access-modes) for more details.

### Query cache information on the host

```shell
$ curl -i http://127.0.0.1:8081/v1/cache/
$ curl -i http://127.0.0.1:8081/v1/cache/l3
$ curl -i http://127.0.0.1:8081/v1/cache/l3/0
```

### Query pre-defined policy in RMD

```shell
$ curl http://127.0.0.1:8081/v1/policy
```

The backend for policy is a YAML file: /etc/rmd/policy.yaml, which pre-defines
some *policies* for target system(Intel platform).

### Create a workload

A workload could be a running task(s) or some set of CPU(s)/Core(s) with cache
allocation information and branch ratio settings.

The task(s)/CPU(s) will be verified, otherwise RMD will fail your request.

Besides you need to specify what policy of the workload will be used.

Depending on available modules the payload body can contain only Cache related params:

```json
{
    "task_ids": [ "A validate task id list" ],
    "core_ids": [ "cpu core list, for the topology, check cache information" ],
    "policy": "pre-defined policy in RMD",
    "cache" : {
        "max": "maximum cache ways which can be benefited",
        "min": "minmum cache ways which can be benefited"
    }
}
```

or also P-State params

```json
{
    "task_ids": [ "A validate task id list" ],
    "core_ids": [ "cpu core list, for the topology, check cache information" ],
    "policy": "pre-defined policy in RMD",
    "cache" : {
        "max": "maximum cache ways which can be benefited",
        "min": "minmum cache ways which can be benefited"
    },
    "pstate" : {
        "ratio": "P-State branch ratio",
        "monitoring": "core(s) monitoring (on or off)"
    }
}
```

For more information about enabling P-State module see [this section](#p-state-module-configuration-and-usage).

Policy (if defined) has higher priority than manually specified params. If policy given
in workload creation request then max_cache, min_cache, pstate_br and monitoring are not
needed (ignored if provided).

An example:

1) Create a workload with gold policy, let say you have a running process
with process id `78377`

```shell
$ curl -H "Content-Type: application/json" --request POST --data \
         '{"task_ids":["78377"],
           "policy": "gold"}' \
         http://127.0.0.1:8081/v1/workloads
```

2) Create workload with manually specified parameters without P-State plugin data.

```shell
$ curl -H "Content-Type: application/json" --request POST --data \
         '{"task_ids" : ["78377"],
           "cache" : {"max": 4, "min": 4 } }' \
         http://127.0.0.1:8081/v1/workloads
```

Please note that it is not allowed (and treated as error) to provide *pstate* parameters section if P-State plugin not enabled in configuration.

3) Create workload with manually specified parameters with P-State plugin enabled

```shell
$ curl -H "Content-Type: application/json" --request POST --data \
         '{"task_ids" : ["78377"],
           "cache" : {"max": 4, "min": 4 },
           "pstate" : {"ratio": 3.0, "monitoring" : "on"} }' \
         http://127.0.0.1:8081/v1/workloads
```

Last param, *monitoring*, can be omitted when defining *ratio* and then the default value "on" will be used.

3) Delete a workload by the workload id, you will find it from the
output of the create response.

```shell
$ curl -H --request DELETE  http://127.0.0.1:8081/v1/workloads/${WORKLOAD_ID}
```

Admin can change and add new policies by editing an toml/yaml file which is
pointed in the configuration file.

```YAML
policypath = "etc/rmd/policy.toml"
```

### Hospitality score API usage:

Hospitality score API will give a score for scheduling workload on a host for
cache allocation request.

Admin can ether give the max_cache/min_cache or a policy to query if the
hospitality score.

The score will be calculate as following:

| request | hospitality score | cache pool |
| :-----: | :---------------: | :--------: |
| max_cache == min_cache > 0 | `[0 \| 100]` | Guarantee |
| max_cache == min_cache == 0 | `[0 \| 100]` | Shared |
| max_cache > min_cache > 0 |  `[0 , 100]` | Besteffort |


To get hospitality score:

```shell
$ curl -H "Content-Type: application/json" --request POST --data \
         '{"max_cache": 2, "min_cache": 2}' \
         http://127.0.0.1:8081/v1/hospitality
{
    "score": {
        "l3": {
            "0": 100,
            "1": 100
        }
    }
}
```

## Supported RMD access modes

### Access RMD by Unix socket:

Access RMD by unit socket if it is enabled.

Requires curl >= v7.40.0
```shell
$ sudo curl --unix-socket /your/socket/path http:/your/resource/url
```

### Access using HTTP requests over plain TCP connection

This method is prepared mainly for RMD development and debugging purposes and is not recommended for use in production system.

For testing purpose, to access REST API with TLS channel disabled, please configure *debugport* param in *debug* section and launch RMD in debug mode:

```shell
$ /path/to/rmd -d
```

or

```shell
$ /path/to/rmd --debug
```

Access to API can done by any HTTP client, like *curl*:

```shell
$ curl -i -X GET http://hostname:debugport/v1/cache
```

### Access using HTTPS over TCP connection secured by TLS:

REST interface of the RMD is recommended to be used over TLS secure channel. RMD supports TLS version 1.2 with cipher suite set to *AES_128_GCM_SHA256* to provide acceptable level of security.

To enable TLS connection, one has to configure *tlsport*, *certpath*, *clientcapath* and *clientauth* options in main configuration file (*rmd.toml*). Also necessary certificates (for client, sever and CA - Certificate authority) have to be provided. Additionally RMD cannot be launched in debug mode (see previous section).

Certificate authority management and certificate generation is out of scope of RMD. For testing purposes, RMD provides pre-defined set of certificates. Please do ***not use these certificates in production environment***.

To access RMD REST API using secure connection use HTTPS protocol and configured *tlsport*. Also provides necessary certificate chain (CA certificate, client certificate and client private key.

Sample command for getting cache info over TLS connection using curl is shown below:

```shell
$ curl https://hostname:tlsport/v1/cache --cert etc/rmd/cert/client/cert.pem \
         --key etc/rmd/cert/client/key.pem \
         --cacert  etc/rmd/cert/client/ca.pem
```
