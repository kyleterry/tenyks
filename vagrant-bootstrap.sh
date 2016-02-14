#!/bin/bash
SRCROOT="/opt/go"
SRCPATH="/opt/gopath"

sudo apt-get update -qq
sudo apt-get upgrade -y
sudo apt-get -q -y install ngircd redis-server wget git mercurial build-essential curl git-core libpcre3-dev pkg-config zip
sudo sed -i 's/127.0.0.1/0.0.0.0/g' /etc/redis/redis.conf
sudo service redis-server restart

# install go I think
cd /tmp
wget --quiet https://storage.googleapis.com/golang/go1.5.3.linux-amd64.tar.gz
tar -xvf go1.5.3.linux-amd64.tar.gz
sudo mv go $SRCROOT
sudo chmod 775 $SRCROOT
sudo chown vagrant:vagrant $SRCROOT
sudo mkdir -p $SRCPATH
sudo chown -R vagrant:vagrant $SRCPATH 2>/dev/null || true
cat <<EOF >/tmp/gopath.sh
export GOPATH="$SRCPATH"
export GOROOT="$SRCROOT"
export PATH="$SRCROOT/bin:$SRCPATH/bin:\$PATH"
export GO15VENDOREXPERIMENT=1
EOF
sudo mv /tmp/gopath.sh /etc/profile.d/gopath.sh
sudo chmod 0755 /etc/profile.d/gopath.sh
source /etc/profile.d/gopath.sh
TENYKS_PATH=${GOPATH}/src/github.com/kyleterry/tenyks
sudo mkdir -p /etc/tenyks
sudo cp "${TENYKS_PATH}/config.json.example" /etc/tenyks/config.json

cd "${TENYKS_PATH}"
go get github.com/githubnemo/CompileDaemon

tee ~vagrant/tmux-launcher.sh <<HERE
tmux -2 new-session -d -s tenyks
tmux new-window -t tenyks -a -n build 'while true; do cd \${GOPATH}/src/github.com/kyleterry/tenyks; CompileDaemon -directory='.' -command='./tenyks'; sleep 2; done'
tmux select-window -t tenyks:1
HERE

bash ~vagrant/tmux-launcher.sh
