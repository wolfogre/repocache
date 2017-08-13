#!/bin/sh

set -e

source /etc/os-release

HOST="{{ . }}"

installEpel(){
    read -t 60 -p "Do you wish to install epel?" YN
    case $YN in
        [Yy]* ) curl -o /etc/yum.repos.d/epel.repo http://$HOST/repo/epel-$VERSION_ID.repo ;;
        [Nn]* ) exit;;
        * ) echo "Please answer yes(Yy) or no(Nn).";;
    esac
}

if [ $ID = "centos" ] ; then
    mv /etc/yum.repos.d/CentOS-Base.repo /etc/yum.repos.d/CentOS-Base.repo.repocache_backup
    if [ $VERSION_ID = "7" ] ; then
        curl -o /etc/yum.repos.d/CentOS-Base.repo http://$HOST/repo/centos-7.repo
        installEpel
    elif [ $VERSION_ID = "6" ] ; then
        curl -o /etc/yum.repos.d/CentOS-Base.repo http://$HOST/repo/centos-6.repo
        installEpel
    else
        mv /etc/yum.repos.d/CentOS-Base.repo.repocache_backup /etc/yum.repos.d/CentOS-Base.repo
        echo "Unsupported centos version: "$VERSION_ID" ."
        echo "Please modify the repo sources manually or contact author to add support."
        exit 1
    fi
    yum makecache
else
    echo "Unsupported linux distribution: "$ID" ."
    echo "Please modify the repo sources manually or contact author to add support."
    exit 1
fi


