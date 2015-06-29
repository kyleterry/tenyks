#!/bin/bash

export GOPATH=~vagrant/go
export PATH=${PATH}:${GOPATH}/bin
echo 'export GOPATH=~vagrant/go' >> .bashrc
echo 'export PATH=${PATH}:${GOPATH}/bin'
mkdir -p "${GOPATH}/src/github.com"
TENYKS_PATH=${GOPATH}/src/github.com/kyleterry/tenyks

sudo chown -R vagrant:vagrant ${GOPATH}

sudo apt-get update -qq
sudo apt-get -q -y install ngircd redis-server wget golang git mercurial
sudo sed -i 's/127.0.0.1/0.0.0.0/g' /etc/redis/redis.conf
sudo service redis-server restart
sudo mkdir -p /etc/tenyks
sudo cp "${TENYKS_PATH}/config.json.example" /etc/tenyks/config.json

cd "${TENYKS_PATH}"
go get github.com/githubnemo/CompileDaemon
go get ./...

tee ~vagrant/tmux-launcher.sh <<HERE
tmux -2 new-session -d -s tenyks
tmux new-window -t tenyks -a -n build 'while true; do cd ~vagrant/go/src/github.com/kyleterry/tenyks; CompileDaemon -directory='.' -command='tenyks'; sleep 2; done'
tmux select-window -t tenyks:1
HERE

bash ~vagrant/tmux-launcher.sh
