// +build openstack

// Package template TODO: This template is too sample. need to improve it.
// Not sure golang native template can satify us.
package template

// Options are options for temlate will be replaced
var Options = map[string]interface{}{
	"address":             "localhost",
	"debugport":           8081,
	"tlsport":             8443,
	"clientauth":          "challenge",
	"logfile":             "/var/log/rmd.log",
	"dbbackend":           "bolt",
	"dbtransport":         "/var/run/rmd.db",
	"logtostdout":         true,
	"os_cacheways":        1,
	"infra_cacheways":     19,
	"max_shared":          10,
	"guarantee":           10,
	"besteffort":          7,
	"shared":              2,
	"shrink":              false,
	"policypath":          "/etc/rmd/policy.toml",
	"sysresctrl":          "/sys/fs/resctrl",
	"plugins":             "",
	"providerConfigPath":  "/etc/nova/provider_config/rmd.yaml",
	"amqpuri":             "amqp://test:test@localhost:5672/",
	"bindingKey":          "versioned_notifications.info",
	"keystoneUrl":         "http://10.237.214.102/identity/v3/auth/tokens",
	"keystoneLogin":       "admin",
	"keystonePassword":    "nimda",
	"dbValidatorInterval": 30, //seconds
	"mbaMode":             "percentage",
}

// Templ is content of template
const Templ = `# This is a rmd config.

title = "RMD Config"

[default]
address = "{{.address}}"
policypath = "{{.policypath}}"
sysresctrl = "{{.sysresctrl}}"
tlsport = {{.tlsport}}
# certpath = "/etc/rmd/cert/server" # Only support pem format, hard code that CAFile is ca.pem, CertFile is rmd-cert.pem, KeyFile is rmd-key.pem
# clientcapath = "/etc/rmd/cert/client" # Only support pem format, hard code that CAFile is ca.pem
clientauth = "{{.clientauth}}"  # can be "no, require, require_any, challenge_given, challenge", challenge means require and verify.
# unixsock = "/var/run/rmd.sock"
plugins = "{{.plugins}}" # comma separated list of enabled RMD plugins, for each plugin (ex. PLUGINX) appropriate config section (ex. [PLUGINX]) is needed
openstackenable = true
dbValidatorInterval = {{.dbValidatorInterval}}

[rdt]
mbaMode = "{{.mbaMode}}" # MBA mode of operation, possible options are: "none", "percentage" and "mbps"

[log]
path = "{{.logfile}}"
env = "dev"  # production or dev
level = "debug"
stdout = {{.logtostdout}}

[database]
backend = "{{.dbbackend}}"
transport = "{{.dbtransport}}"
dbname = "bolt"

[debug]
enabled = false
debugport = {{.debugport}}

[OSGroup] # OSGroup is mandatory
cacheways = {{.os_cacheways}}
cpuset = "0-1"

[InfraGroup] # InfraGroup is optional
cacheways = {{.infra_cacheways}}
cpuset = "2-3"
tasks = ["ovs*"] # Just support Wildcards. Do we need to support RE?

[CachePool] # Cache Pool config is optional
shrink = {{.shrink}}
max_allowed_shared = {{.max_shared}} # max allowed workload in shared pool, default is {{.max_shared}}
guarantee = {{.guarantee}}
besteffort = {{.besteffort}}
shared = {{.shared}}

[acl]
path = "/etc/rmd/acl/"  #
# use CSV format
filter = "url" # at present just support "url", will support "ip, proto"
authorization = "signature" # authorize the client, can identify client by signature, role(OU) or username(CN). Default value is signature. If value is signature, admincert and usercert should be set.
admincert = "/etc/rmd/acl/roles/admin/" # A cert is used to describe user info. These cert files in this path are used to define the users that are admin. Only pem format file at present. The files can be updated dynamicly.
usercert = "/etc/rmd/acl/roles/user/" # A cert is used to describe user info. These cert files in this path are used to define the user with low privilege. Only pem format file at present. The files can be updated dynamicly.

[pam]
service = "rmd"

[openstack]
# Path below is optional. If not given then file will not be generated
#providerConfigPath = "{{.providerConfigPath}}"
#amqpuri = "{{.amqpuri}}" # URI to openstack amqp server, required to subscribe to nova notifications
#bindingKey = "{{.bindingKey}}"
#keystoneUrl = "{{.keystoneUrl}}"  # URI to Keystone server, required to get token
#keystoneLogin = "{{.keystoneLogin}}"
#keystonePassword = "{{.keystonePassword}}"
`
