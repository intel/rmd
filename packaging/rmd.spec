Name:           rmd
Version:        1.0
Release:        1%{?dist}
Summary:        Resource Management Daemon-RMD
License:        ASL 2.0
URL:            https://github.com/intel/rmd
Source0:        https://www.example.com/%{name}/releases/%{name}-%{version}.tar.gz
BuildRequires:  go
BuildRequires:  make
BuildRequires:  pam-devel
# this package does not support big endian arch so far,
# and has been verified only on Intel platforms.
ExclusiveArch: %{ix86} x86_64


%description
RMD application

%prep
%setup -q

%build
make %{?_smp_mflags}

%install
mkdir -p %{buildroot}/%{_bindir}/
install -p -m 755 %{_builddir}/%{name}-%{version}/build/linux/x86_64/rmd %{buildroot}/%{_bindir}/
install -p -m 755 %{_builddir}/%{name}-%{version}/build/linux/x86_64/gen_conf %{buildroot}/%{_bindir}/
mkdir -p %{buildroot}/%{_bindir}/scripts
cp -r  %{_builddir}/%{name}-%{version}/scripts/* %{buildroot}/%{_bindir}/scripts
mkdir -p %{buildroot}/%{_bindir}/etc/rmd
cp -r %{_builddir}/%{name}-%{version}/etc/rmd %{buildroot}/%{_bindir}/etc


%files
%{_bindir}/%{name}
%{_bindir}/gen_conf
%config(missingok) %{_bindir}/scripts
%config(missingok) %{_bindir}/etc
%doc README.md
%license LICENSE

%post
%{_bindir}/scripts/install.sh --skip-pam-userdb
rm -rf %{_bindir}/scripts
rm -rf %{_bindir}/etc

%preun
rm -rf /etc/rmd/pam/rmd_users.db
rm -rf /usr/local/etc/rmd
rm -rf /var/run/rmd

%changelog
* Tue Jan 07 2020 ArunPrabhu Vijayan <arunprabhu.vijayan@intel.com> - 1.0-1
- RMD package version 1.0
