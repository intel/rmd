# SamplePlugin user guide

SamplePlugin is a dummy RMD plugin that can be used as a reference or template for plugin development.

## Building, configuring and launching

SamplePlugin is a simple Go project to be compiled into loadable *.so* file. It's only dependency is [Go-Restful](github.com/emicklei/go-restful) project used for handling REST request. Please note that it doesn't depend on any RMD component so it can be build completely out of RMD repository using *make*

```shell
make
make install
```

Configuration and launching of this plugin consist of following steps:

1. Add plugin to the list of enabled plugins in *rmd.toml* (section: [default] parameter: *plugins*)
  ```toml
  [default]
  # ...
  # ...
  plugins = "sampleplugin"
  ```
2. Add new configuration section [sampleplugin] with plugin configuration
  ```toml
  [sampleplugin]
  # path where compiled library is stored
  path = "/etc/rmd/plugins/sampleplugin.so"
  # parameter used to show how initialization works, any positive integer value is OK
  identifier = 1234
  ```
3. Restart RMD (if running)

## Using REST API endpoint

This SamplePlugin provides two REST endpoints: */sampleplugin* and */sampleplugin/status*. First one returns overall plugin information (name hardcoded in implementation, and identifier from config file). Second endpoint returns only a plugins' status (one boolean, should be true in properly working and initialized module). As stated in RMD *PluginDevelopmentGuide* these endpoints are prepared only to handle GET requests.

Below please find sample REST calls with returned results:

```
$> curl -X GET http://127.0.0.1:8081/v1/sampleplugin
{
  "name": "sampleplugin",
  "id": 1234
 }[

$> curl -X GET http://127.0.0.1:8081/v1/sampleplugin/status
{
  "status": true
 }
```

## Using inside RMD workload

This plugin does not access any platform resources to make it's implementation as simple as possible but still can be used as a part or RMD *Workload*.

To create *workload* with *sampleplugin* usage please send POST request with *plugins* section containing parameters for *sampleplugin*. To show how param validation works two parameters were introduced:

* *name* - parameter of type string, cannot contain spaces and tabulation characters (as this will fail validation), optional
* *value* - real number (with a decimal point), mandatory

Please note that these parameters have no impact on *Enforce()* function execution - they're just verified and rejected if invalid. Below please find sample POST requests:


* with valid set of params (all params defined)
```
$> curl -X POST -H "Content-Type: application/json" --data '{
    "core_ids":["10"], "plugins" : {
        "sampleplugin" : {
            "name" : "somestring", "value" : 12.5
            }
        }
    }'  http://127.0.0.1:8081/v1/workloads
```
* valid set of params (only mandatory param defined)
```
$> curl -X POST -H "Content-Type: application/json" --data '{
    "core_ids":["10"], "plugins" : {
        "sampleplugin" : {
            "name" : "somestring", "value" : 12.5
            }
        }
    }'  http://127.0.0.1:8081/v1/workloads
```
* invalid set of params (invalid type)
```
$> curl -X POST -H "Content-Type: application/json" --data '{
    "core_ids":["10"], "plugins" : {
        "sampleplugin" : {
            "name" : "somestring", "value" : "1234"
            }
        }
    }'  http://127.0.0.1:8081/v1/workloads
```
* invalid set of params (missing mandatory param)
```
$> curl -X POST -H "Content-Type: application/json" --data '{
    "core_ids":["10"], "plugins" : {
        "sampleplugin" : {
            "name" : "somestring"
            }
        }
    }'  http://127.0.0.1:8081/v1/workloads
```

