#!/bin/bash

if [ `id -u` != "0" ]; then
    echo "You must be root to run this script."
    exit 1
fi

usage() {
    echo "Usage: ${0} [distro]"
}

if [ $# -eq 0 ]; then
    usage
    exit 1
fi

cat <<HERE
  _                   _         
 | |                 | |        
 | |_ ___ _ __  _   _| | _____  
 | __/ _ \ '_ \| | | | |/ / __| 
 | ||  __/ | | | |_| |   <\__ \  irc bot install.
  \__\___|_| |_|\__, |_|\_\___/ 
                 __/ |          
                |___/           

HERE


_script="$(readlink -f ${BASH_SOURCE[0]})"
 
BASE_DIR="$(dirname $_script)"


TENYKS_BIN=`which tenyks`
if [ ! -z ${TENYKS_BIN} ]; then
    printf "Found tenyks in ${TENYKS_BIN}\n\n"
else
    echo "You must install tenyks before running this script"
    exit 1
fi

#_START="yes"
#PS3="> "
#echo "Should I try to start Tenyks when we finish?"
#select START in "yes" "no"
#do
#    case ${START} in
#        yes)
#            break
#            ;;
#        no)
#            break
#            ;;
#    esac
#done

_USER=tenyks
read -p "Who is running tenyks? [${_USER}]" USER
if [ !"${USER}" ]; then
    USER=${_USER}
    EXISTS=$(cat /etc/passwd |grep ${USER} | awk -F : '{print $1}')
    if [ -z ${EXISTS} ]; then
        USER_EXISTS=false
        echo "Should I create the user? "
        select CREATE_USER in "yes" "no"
        do
            case ${CREATE_USER} in
                yes)
                    CREATE_USER=true
                    break
                    ;;
                no)
                    CREATE_USER=false
                    break
                    ;;
            esac
        done
    else
        USER_EXISTS=true
    fi
fi

case "$1" in
    ubuntu)
        SETTINGS_DIR="/etc/tenyks"
        SETTINGS_PATH="${SETTINGS_DIR}/settings.py"
        LOG_FILE="/var/log/tenyks.log"

        printf "Installing Tenyks Ubuntu style...\n\n"

        echo "Installing init script..."
        cp ${BASE_DIR}/tenyks_init /etc/init.d/tenyks
        update-rc.d tenyks defaults

        echo "Making settings..."
        mkdir -p ${SETTINGS_DIR}
        tenyksmkconfig > ${SETTINGS_PATH}

        echo "Touching file..."
        touch ${LOG_FILE}
        echo "Changing owner to ${USER}..."
        chown ${USER} ${LOG_FILE}

        echo "Okay done. You should edit your settings in ${SETTINGS_PATH}"
        echo "Start Tenyks with: service tenyks start"
        ;;
    arch)
        echo "Installing Tenyks Arch Linux style..."
        ;;
    *)
        echo "Pull requests welcome :)"
        exit 1
        ;;
esac

