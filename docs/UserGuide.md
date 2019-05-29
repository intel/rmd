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

RMD will try to search `/usr/local/etc/rmd`, `/etc/rmd`, `./etc/rmd` to find
configure files, put configuration files into these directory or RMD will
use default configurations.

Here's a sample for the main configure file of RMD.

Sample configuration file: (/usr/local/etc/rmd/rmd.toml)
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
    $ sudo ./scripts/setup_rmd_users.sh
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
$ sudo rmd --conf-dir /usr/local/etc/rmd
```

RMD can be launched in debug mode that exposes RESTAPI service on HTTP.
By default it is launched on port 8081. (http://127.0.0.1:8081)

```shell
$ sudo rmd --debug
```

## RMD service usages

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

A workload could be a running task(s) or some CPU(s) which want to allocate
cache for them.

The task(s)/CPU(s) will be verified, otherwise RMD will fail your request.

Besides you need to specify what policy of the workload will be used.

The payload body can contain:

```json
{
    "task_ids": A validate task id list
    "core_ids": cpu core list, for the topology, check cache information
    "policy": pre-defined policy in RMD
    "max_cache": maximum cache ways which can be benefited
    "min_cache": minmum cache ways which can be benefited
}
```

You can not neither specify policy and max_cache/min_cache at same time, that
is ambiguous to RMD.

An example:

1) Create a workload with gold policy, let say you have a running process
with process id `78377`

```shell
$ curl -H "Content-Type: application/json" --request POST --data \
         '{"task_ids":["78377"], "policy": "gold"}' \
         http://127.0.0.1:8081/v1/workloads
```

2) Create workload with max_cache, min_cache.

```shell
$ curl -H "Content-Type: application/json" --request POST --data \
         '{"task_ids":["78377"], "max_cache": 4, "min_cache": 4}' \
         http://127.0.0.1:8888/v1/workloads
```

3) Delete a workload by the workload id, you will find it from the
output of the create response.

```shell
$ curl -H "Content-Type: application/json" \
        --request DELETE  http://127.0.0.1:8888/v1/workloads/${WORKLOAD_ID}
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
         http://127.0.0.1:8888/v1/hospitality
{
    "score": {
        "l3": {
            "0": 100,
            "1": 100
        }
    }
}
```

### Access RMD by Unix socket:

Access RMD by unit socket if it is enabled.

Requires curl >= v7.40.0
```shell
$ sudo curl --unix-socket /your/socket/path http:/your/resource/url
```

### Access RMD by TLS:

Access RMD by TLS if it is enabled.

Need to config tlsport, certpath, clientcapath, clientauth options in
configure file.

Using TLS and managing a CA is an advanced topic. It is not the scope of RMD.
RMD just pre-define server certs for testing.

Please do not use them in product environment.
User can generate certs by themselves.

If you want to get cache info, your can run this command:
```shell
$ curl https://hostname:port/v1/cache --cert etc/rmd/cert/client/cert.pem \
         --key etc/rmd/cert/client/key.pem \
         --cacert  etc/rmd/cert/client/ca.pem
```
