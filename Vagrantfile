# -*- mode: ruby -*-
# vi: set ft=ruby :

# Vagrantfile API/syntax version. Don't touch unless you know what you're doing!
VAGRANTFILE_API_VERSION = "2"

$script = <<SCRIPT
apt-get update -qq
apt-get -q -y install ngircd redis-server wget golang git
mkdir -p /tmp/tenyks
cd /tmp/tenyks
cp -r /vagrant/* /tmp/tenyks
make
make install
make clean
mkdir -p /etc/tenyks
cp config.json.example /etc/tenyks/config.json
SCRIPT

Vagrant.configure(VAGRANTFILE_API_VERSION) do |config|
  config.vm.box = "ubuntu/trusty64"
  config.vm.network "private_network", ip: "192.168.33.66"
  config.vm.hostname = "tenyks"
  config.vm.provision "shell", inline: $script
end
