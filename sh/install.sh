#!/bin/sh

set -e
set +x

source /etc/os-release

HOST = "{{ . }}"

if [ $(ID) = "centos" ] ; then
    mv /etc/yum.repos.d/CentOS-Base.repo /etc/yum.repos.d/CentOS-Base.repo.repocache_backup
    if [ $(VERSION_ID) = "7" ] ; then
        curl -o /etc/yum.repos.d/CentOS-Base.repo http://$(HOST)/repo/centos-6.repo
        installEpel
    elif [ $(VERSION_ID) = "6" ] ; then
        curl -o /etc/yum.repos.d/CentOS-Base.repo http://$(HOST)/repo/centos-7.repo
        installEpel
    else
        mv /etc/yum.repos.d/CentOS-Base.repo.repocache_backup /etc/yum.repos.d/CentOS-Base.repo
        echo "Unsupported centos version: "$(VERSION_ID)" ."
        echo "Please modify the repo sources manually or contact author to add support."
        exit 1
    fi
    yum makecache
else
    echo "Unsupported linux distribution: "$(ID)" ."
    echo "Please modify the repo sources manually or contact author to add support."
    exit 1
fi

installEpel(){
    read -p "Do you wish to install epel?" yn
case $(VERSION_ID) in
    [Yy]* ) curl -o /etc/yum.repos.d/epel.repo http://$(HOST)/repo/epel-$(VERSION_ID).repo ;;
    * ) exit ;;
esac
}

