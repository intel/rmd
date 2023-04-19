%global goipath github.com/intel/rmd

Name:           rmd
Version:        0.3.1
Release:        1%{?dist}
Summary:        Resource Management Daemon-RMD
License:        ASL 2.0 and  BSD and MIT and MPLv2.0 
URL:            https://github.com/intel/rmd
Source0:        https://github.com/intel/rmd/archive/v0.3.1.tar.gz

BuildRequires:  go
BuildRequires:  make
BuildRequires:  pam-devel
BuildRequires:  systemd
BuildRequires:  systemd-rpm-macros
BuildRequires:  git-core
#intel-cmt-cat-devel :BSD
Requires:  intel-cmt-cat-devel >= 2.0.0-3
#github.com/Knetic/govaluate : MIT
Provides:      bundled(golang(github.com/Knetic/govaluate)) = 3.0.1
#github.com/bgentry/speakeasy : MIT
Provides:      bundled(golang(github.com/bgentry/speakeasy)) =	0.1.0
#github.com/casbin/casbin : ASL 2.0
Provides:      bundled(golang(github.com/casbin/casbin)) = 1.9.1
#github.com/fatih/structs : MIT
Provides:      bundled(golang(github.com/fatih/structs)) = 1.1.0
#github.com/globalsign/mgo : BSD
Provides:      bundled(golang(github.com/globalsign/mgo)) = 0.0.0
#github.com/gobwas/glob : MIT
Provides:      bundled(golang(github.com/gobwas/glob)) = 0.2.3
#github.com/gopherjs/gopherjs : BSD
Provides:      bundled(golang(github.com/gopherjs/gopherjs)) = 0.0.0
#github.com/hashicorp/hcl : MPLv2.0
Provides:      bundled(golang(github.com/hashicorp/hcl)) = 1.0.0
#github.com/jtolds/gls : MIT
Provides:      bundled(golang(github.com/jtolds/gls)) = 4.20.0
#github.com/klauspost/compress : BSD
Provides:      bundled(golang(github.com/klauspost/compress)) = 1.10.6
#github.com/klauspost/cpuid : MIT
Provides:      bundled(golang(github.com/klauspost/cpuid)) = 1.10.6
#github.com/kr/pretty : MIT
Provides:      bundled(golang(github.com/kr/pretty)) = 0.1.0
#github.com/magiconair/properties : BSD
Provides:      bundled(golang(github.com/magiconair/properties)) = 1.8.1
#github.com/mitchellh/mapstructure : MIT
Provides:      bundled(golang(github.com/mitchellh/mapstructure)) = 1.1.2
#github.com/onsi/ginkgo : MIT
Provides:      bundled(golang(github.com/onsi/ginkgo)) = 1.14.2
#github.com/onsi/gomega : MIT
Provides:      bundled(golang(github.com/onsi/gomega)) = 1.10.1
#github.com/sirupsen/logrus : MIT
Provides:      bundled(golang(github.com/sirupsen/logrus)) = 1.6.0
#github.com/spf13/afero : ASL 2.0
Provides:      bundled(golang(github.com/spf13/afero)) = 1.1.2
#github.com/spf13/cast : MIT
Provides:      bundled(golang(github.com/spf13/cast)) = 1.3.0
#github.com/spf13/jwalterweatherman : MIT
Provides:      bundled(golang(github.com/spf13/jwalterweatherman)) = 1.0.0
#github.com/spf13/pflag : BSD
Provides:      bundled(golang(github.com/spf13/pflag)) = 1.0.5
#github.com/spf13/viper : MIT
Provides:      bundled(golang(github.com/spf13/viper)) = 1.7.0
#github.com/streadway/amqp : BSD
Provides:      bundled(golang(github.com/streadway/amqp)) = 1.0.0
#github.com/stretchr/testify : MIT
Provides:      bundled(golang(github.com/stretchr/testify)) = 1.3.0
#github.com/valyala/bytebufferpool : MIT
Provides:      bundled(golang(github.com/valyala/bytebufferpool)) = 1.0.0
#github.com/xeipuuv/gojsonschema : ASL 2.0
Provides:      bundled(golang(github.com/xeipuuv/gojsonschema)) = 1.2.0
#github.com/yudai/gojsondiff : MIT
Provides:      bundled(golang(github.com/yudai/gojsondiff)) = 1.0.0
#github.com/yudai/golcs : MIT
Provides:      bundled(golang(github.com/yudai/golcs)) = 0.0.0
#golang.org/x/sys/cpu : BSD
Provides:      bundled(golang(golang.org/x/sys/cpu)) = 0.0.0
#gopkg.in/yaml.v2 : ASL 2.0 and MIT
Provides:      bundled(golang(gopkg.in/yaml.v2)) = 2.3.0
#github.com/golang/glog : ASL 2.0
Provides:      bundled(golang(github.com/golang/glog)) = 0.0.0
#github.com/etcd-io/bbolt : MIT
Provides:      bundled(golang(github.com/etcd-io/bbolt)) = 1.3.3
#gopkg.in/yaml.v2 : ASL 2.0 and MIT
Provides:	bundled(golang(gopkg.in/yaml.v2)) = 2.3.0
#github.com/streadway/amqp : BSD
Provides:	bundled(golang(github.com/streadway/amqp)) = 1.0.0

# this package does not support big endian arch so far,
# and has been verified only on Intel platforms.
# HW support is documented in https://github.com/intel/rmd/blob/master/docs/Prerequisite.md
ExclusiveArch: %{ix86} x86_64

%description
RMD is a system daemon providing a central interface for
hardware resource management tasks on x86 platforms.

%prep
%setup -q

%build
export GOPATH=${PWD}
export PATH=${GOPATH}:${PATH}
rsync -az --exclude=gopath/ ./ %{_builddir}/%{name}-%{version}
cd %{_builddir}/%{name}-%{version}
make %{?_smp_mflags} VERSION=${RMD_VERSION}

%install

mkdir -p %{buildroot}/%{_bindir}/
GOOS=${GOOS:-$(go env GOOS)}
GOARCH=${GOARCH:-$(go env GOARCH)}
if [[ "${GOARCH}" == "amd64" ]]; then
    GOARCH="x86_64"
fi

install -p -m 755 %{_builddir}/%{name}-%{version}/build/$GOOS/$GOARCH/rmd %{buildroot}%{_bindir}/
install -p -m 755 %{_builddir}/%{name}-%{version}/build/$GOOS/$GOARCH/gen_conf %{buildroot}%{_bindir}/

install -d %{buildroot}%{_mandir}/man8
install -m 0644  %{_builddir}/%{name}-%{version}/rmd.8 %{buildroot}%{_mandir}/man8
ln -sf %{_mandir}/man8/rmd.8 %{buildroot}%{_mandir}/man8/gen_conf.8

mkdir -p %{buildroot}%{_datadir}/%{name}/scripts
install -m 755  %{_builddir}/%{name}-%{version}/scripts/setup_rmd_users.sh %{buildroot}%{_datadir}/%{name}/scripts

mkdir -p %{buildroot}%{_unitdir}
install -m 644 %{_builddir}/%{name}-%{version}/scripts/%{name}.service %{buildroot}%{_unitdir}

mkdir -p %{buildroot}%{_sysconfdir}/rmd
install -m 0644  %{_builddir}/%{name}-%{version}/etc/rmd/cpu_map.toml %{buildroot}%{_sysconfdir}/rmd
install -m 0644  %{_builddir}/%{name}-%{version}/etc/rmd/policy.toml %{buildroot}%{_sysconfdir}/rmd
install -m 0644  %{_builddir}/%{name}-%{version}/etc/rmd/policy.yaml %{buildroot}%{_sysconfdir}/rmd
install -m 0644  %{_builddir}/%{name}-%{version}/etc/rmd/rmd.toml %{buildroot}%{_sysconfdir}/rmd

mkdir -p %{buildroot}%{_sysconfdir}/rmd/acl/roles/admin
install -m 0644  %{_builddir}/%{name}-%{version}/etc/rmd/acl/roles/admin/cert.pem %{buildroot}%{_sysconfdir}/rmd/acl/roles/admin

mkdir -p %{buildroot}%{_sysconfdir}/rmd/acl/roles/user
install -m 0644  %{_builddir}/%{name}-%{version}/etc/rmd/acl/roles/user/user-cert.pem %{buildroot}%{_sysconfdir}/rmd/acl/roles/user

mkdir -p %{buildroot}%{_sysconfdir}/rmd/acl/url
install -m 0644  %{_builddir}/%{name}-%{version}/etc/rmd/acl/url/model.conf %{buildroot}%{_sysconfdir}/rmd/acl/url
install -m 0644  %{_builddir}/%{name}-%{version}/etc/rmd/acl/url/policy.csv %{buildroot}%{_sysconfdir}/rmd/acl/url

mkdir -p %{buildroot}%{_sysconfdir}/rmd/cert/client
install -m 0644  %{_builddir}/%{name}-%{version}/etc/rmd/cert/client/ca.pem %{buildroot}%{_sysconfdir}/rmd/cert/client
install -m 0644  %{_builddir}/%{name}-%{version}/etc/rmd/cert/client/cert.pem %{buildroot}%{_sysconfdir}/rmd/cert/client
install -m 0644  %{_builddir}/%{name}-%{version}/etc/rmd/cert/client/key.pem %{buildroot}%{_sysconfdir}/rmd/cert/client
install -m 0644  %{_builddir}/%{name}-%{version}/etc/rmd/cert/client/user-cert.pem %{buildroot}%{_sysconfdir}/rmd/cert/client
install -m 0644  %{_builddir}/%{name}-%{version}/etc/rmd/cert/client/user-key.pem %{buildroot}%{_sysconfdir}/rmd/cert/client

mkdir -p %{buildroot}%{_sysconfdir}/rmd/cert/server
install -m 0644  %{_builddir}/%{name}-%{version}/etc/rmd/cert/server/ca.pem %{buildroot}%{_sysconfdir}/rmd/cert/server
install -m 0644  %{_builddir}/%{name}-%{version}/etc/rmd/cert/server/rmd-cert.pem %{buildroot}%{_sysconfdir}/rmd/cert/server
install -m 0644  %{_builddir}/%{name}-%{version}/etc/rmd/cert/server/rmd-key.pem %{buildroot}%{_sysconfdir}/rmd/cert/server

mkdir -p %{buildroot}%{_sysconfdir}/rmd/pam
install -m 0644  %{_builddir}/%{name}-%{version}/etc/rmd/pam/rmd %{buildroot}%{_sysconfdir}/rmd/pam

mkdir -p %{buildroot}%{_sysconfdir}/rmd/pam/test
install -m 0644  %{_builddir}/%{name}-%{version}/etc/rmd/pam/test/rmd %{buildroot}%{_sysconfdir}/rmd/pam/test

mkdir -p %{buildroot}%{_docdir}/%{name}
install -m 0644  %{_builddir}/%{name}-%{version}/docs/UserGuide.md %{buildroot}%{_docdir}/rmd
install -m 0644  %{_builddir}/%{name}-%{version}/docs/Prerequisite.md %{buildroot}%{_docdir}/rmd
install -m 0644  %{_builddir}/%{name}-%{version}/docs/ConfigurationGuide.md %{buildroot}%{_docdir}/rmd

%files
%{_bindir}/%{name}
%{_bindir}/gen_conf
%{_mandir}/man8/rmd.8.*
%{_mandir}/man8/gen_conf.8.*
%{_datadir}/%{name}/
%config(noreplace)  %{_sysconfdir}/rmd/cert/*
%config(noreplace)  %{_sysconfdir}/rmd/acl/*
%config(noreplace)  %{_sysconfdir}/rmd/*.toml
%config(noreplace)  %{_sysconfdir}/rmd/*.yaml
%config(noreplace)  %{_sysconfdir}/rmd/pam/test/rmd
%config(noreplace)  %{_sysconfdir}/rmd/pam/rmd
%doc README.md CHANGELOG.md ./docs/UserGuide.md ./docs/Prerequisite.md ./docs/ConfigurationGuide.md
%license LICENSE

%{_unitdir}/%{name}.service


%post

%systemd_post %{name}.service

USER="rmd"
useradd $USER || echo "User rmd already exists."

LOGFILE="/var/log/rmd/rmd.log"
if [ ! -d ${LOGFILE%/*} ]; then
    mkdir -p ${LOGFILE%/*}
    chown $USER:$USER ${LOGFILE%/*}
fi

DBFILE="/var/run/rmd/rmd.db"
if [ ! -d ${DBFILE%/*} ]; then
    mkdir -p ${DBFILE%/*}
    chown $USER:$USER  ${DBFILE%/*}
fi

PAMSRCFILE="/etc/rmd/pam/rmd"
PAMDIR="/etc/pam.d"
if [ -d $PAMDIR ]; then
    cp $PAMSRCFILE $PAMDIR
fi

DATA="\"logfile\":\"$LOGFILE\", \"dbtransport\":\"$DBFILE\", \"logtostdout\":false"
gen_conf -path /etc/rmd/rmd.toml -data "{$DATA}"

%preun 
%systemd_preun %{name}.service
USER="rmd"
rm -rf /var/log/rmd/
rm -rf /var/run/rmd/
rm -rf /etc/pam.d/rmd

%changelog
* Thu Jan 21 2021 Gargi Sau <gargi.sau@intel.com> - 0.3.1-1
- RMD package version 0.3.1

* Mon Jun 22 2020 ArunPrabhu Vijayan <arunprabhu.vijayan@intel.com> - 0.3-1
- New release 0.3

* Thu Jun 04 2020 ArunPrabhu Vijayan <arunprabhu.vijayan@intel.com> - 0.2.1-1
- RMD package version 0.2.1
