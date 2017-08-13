#!/bin/sh

set -e
set +x

source /etc/os-release

if [ $(ID) = "centos" ] ; then
    mv /etc/yum.repos.d/CentOS-Base.repo.repocache_backup /etc/yum.repos.d/CentOS-Base.repo
    yum makecache
else
    echo "Unsupported linux distribution: "$ID" ."
    echo "Please modify the repo sources manually or contact author to add support."
    exit 1
fi