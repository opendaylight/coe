
Vagrant.configure(2) do |config|
  config.vm.box = "bento/ubuntu-16.04"

  # Spawn up the k8 master node
  config.vm.define "k8se2emaster" do |k8se2emaster|
    k8se2emaster.vm.hostname = "k8se2emaster"   
    k8se2emaster.vm.network "private_network", ip: "192.168.30.11"
    k8se2emaster.vm.network "private_network", ip: "192.168.40.11"
    k8se2emaster.vm.provision "shell", path: "provisioning/setup-Req.sh", privileged: false
    k8se2emaster.vm.provision "shell", path: "provisioning/setup-k8s-master.sh", privileged: false
    k8se2emaster.vm.provider "virtualbox" do |vb|
       vb.name = "k8se2emaster"
       vb.customize ["modifyvm", :id, "--memory", "2048"]
    end
  end

  # Spawn up the k8 minion1 node
  config.vm.define "k8se2eminion1" do |k8se2eminion1|
    k8se2eminion1.vm.hostname = "k8se2eminion1"
    k8se2eminion1.vm.network "private_network", ip: "192.168.30.12"
    k8se2eminion1.vm.network "private_network", ip: "192.168.40.12"
    k8se2eminion1.vm.provision "shell", path: "provisioning/setup-Req.sh", privileged: false
    k8se2eminion1.vm.provider "virtualbox" do |vb|
       vb.name = "k8se2eminion1"
       vb.customize ["modifyvm", :id, "--memory", "2048"]
    end
  end

  # Spawn up the k8 minion2 node
  config.vm.define "k8se2eminion2" do |k8se2eminion2|
    k8se2eminion2.vm.hostname = "k8se2eminion2"
    k8se2eminion2.vm.network "private_network", ip: "192.168.30.13"
    k8se2eminion2.vm.network "private_network", ip: "192.168.40.13"
    k8se2eminion2.vm.provision "shell", path: "provisioning/setup-Req.sh", privileged: false
    k8se2eminion2.vm.provider "virtualbox" do |vb|
       vb.name = "k8se2eminion2"
       vb.customize ["modifyvm", :id, "--memory", "2048"]
    end
  end

  config.vm.provider "virtualbox" do |v|
    v.customize ["modifyvm", :id, "--cpuexecutioncap", "50"]
  end
end
