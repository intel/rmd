How do I build and install from a RPM Spec file?

Basic steps to build source and binary packages in your home directory:

Create the rpmbuild directory structure:
$ rpmdev-setuptree

Next, download the RMD tar file and copy into ~/rpmbuild/SOURCES directory:
$ wget https://github.com/intel/rmd/archive/v0.3.1.tar.gz 
$ cp ./v0.3.1.tar.gz ~/rpmbuild/SOURCES

To build, do:
$ rpmbuild -ba path/to/rmd.spec

To install, do:
yum install ~/rpmbuild/RPMS/x86_64/rmd-0.3.1-1.fc32.x86_64.rpm

