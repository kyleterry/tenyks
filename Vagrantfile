# -*- mode: ruby -*-
# vi: set ft=ruby :

# Vagrantfile API/syntax version. Don't touch unless you know what you're doing!
VAGRANTFILE_API_VERSION = "2"

Vagrant.configure(VAGRANTFILE_API_VERSION) do |config|
  config.vm.box = "ubuntu/trusty64"
  config.vm.hostname = "tenyks"
  config.vm.provision "shell", path: './vagrant-bootstrap.sh', privileged: false
  config.vm.network "forwarded_port", guest: 6667, host: 6667
  config.vm.network "forwarded_port", guest: 6379, host: 6378
  config.vm.synced_folder ".", "/home/vagrant/go/src/github.com/kyleterry/tenyks"
end
