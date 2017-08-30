#!/bin/sh

set -e

command_exists() {
	command -v "$@" > /dev/null 2>&1
}

# perform some very rudimentary platform detection
get_version() {
    lsb_dist=''
    dist_version=''
    if command_exists lsb_release; then
        lsb_dist="$(lsb_release -si)"
    fi
    if [ -z "$lsb_dist" ] && [ -r /etc/lsb-release ]; then
        lsb_dist="$(. /etc/lsb-release && echo "$DISTRIB_ID")"
    fi
    if [ -z "$lsb_dist" ] && [ -r /etc/debian_version ]; then
        lsb_dist='debian'
    fi
    if [ -z "$lsb_dist" ] && [ -r /etc/fedora-release ]; then
        lsb_dist='fedora'
    fi
    if [ -z "$lsb_dist" ] && [ -r /etc/oracle-release ]; then
        lsb_dist='oracleserver'
    fi
    if [ -z "$lsb_dist" ]; then
        if [ -r /etc/centos-release ] || [ -r /etc/redhat-release ]; then
            lsb_dist='centos'
        fi
    fi
    if [ -z "$lsb_dist" ] && [ -r /etc/os-release ]; then
        lsb_dist="$(. /etc/os-release && echo "$ID")"
    fi

    lsb_dist="$(echo "$lsb_dist" | tr '[:upper:]' '[:lower:]')"

    case "$lsb_dist" in

        ubuntu)
            if command_exists lsb_release; then
                dist_version="$(lsb_release --codename | cut -f2)"
            fi
            if [ -z "$dist_version" ] && [ -r /etc/lsb-release ]; then
                dist_version="$(. /etc/lsb-release && echo "$DISTRIB_CODENAME")"
            fi
        ;;

        debian)
            dist_version="$(cat /etc/debian_version | sed 's/\/.*//' | sed 's/\..*//')"
            case "$dist_version" in
                8)
                    dist_version="jessie"
                ;;
                7)
                    dist_version="wheezy"
                ;;
            esac
        ;;

        oracleserver)
            # need to switch lsb_dist to match yum repo URL
            lsb_dist="oraclelinux"
            dist_version="$(rpm -q --whatprovides redhat-release --queryformat "%{VERSION}\n" | sed 's/\/.*//' | sed 's/\..*//' | sed 's/Server*//' | sort -r | head -1)"
        ;;

        fedora|centos)
            dist_version="$(rpm -q --whatprovides redhat-release --queryformat "%{VERSION}\n" | sed 's/\/.*//' | sed 's/\..*//' | sed 's/Server*//' | sort -r | head -1)"
        ;;

        *)
            if command_exists lsb_release; then
                dist_version="$(lsb_release --codename | cut -f2)"
            fi
            if [ -z "$dist_version" ] && [ -r /etc/os-release ]; then
                dist_version="$(. /etc/os-release && echo "$VERSION_ID")"
            fi
        ;;
    esac
}

HOST="{{ . }}"

installEpel(){
    read -t 60 -p "Do you wish to install epel?" YN
    case $YN in
        [Yy]* ) curl -o /etc/yum.repos.d/epel.repo http://$HOST/repo/epel-$dist_version.repo ;;
        [Nn]* ) exit;;
        * ) echo "Please answer yes(Yy) or no(Nn).";;
    esac
}

if [ "$lsb_dist" = "centos" ] ; then
    mv /etc/yum.repos.d/CentOS-Base.repo /etc/yum.repos.d/CentOS-Base.repo.repocache_backup
    if [ $dist_version = "7" ] ; then
        curl -o /etc/yum.repos.d/CentOS-Base.repo http://$HOST/repo/centos-7.repo
        installEpel
    elif [ $dist_version = "6" ] ; then
        curl -o /etc/yum.repos.d/CentOS-Base.repo http://$HOST/repo/centos-6.repo
        installEpel
    else
        mv /etc/yum.repos.d/CentOS-Base.repo.repocache_backup /etc/yum.repos.d/CentOS-Base.repo
        echo "Unsupported centos version: "$dist_version" ."
        echo "Please modify the repo sources manually or contact author to add support."
        exit 1
    fi
    yum makecache
else
    echo "Unsupported linux distribution: "$lsb_dist" ."
    echo "Please modify the repo sources manually or contact author to add support."
    exit 1
fi


