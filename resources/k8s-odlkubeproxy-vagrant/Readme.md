**Steps:**

- Run vagrant up to spawn the VMs $ vagrant up

- For each vm do the following:

  - ssh to the vm using the following command:
    - $ vagrant ssh k8se2emaster
    - vagrant@k8smaster$ sudo mkdir -p /etc/cni/net.d/

  - Copy the odlOvSCNI binary exist under /vagrant/cni_conf/ to /opt/cni/bin
     - vagrant@k8smaster$ sudo cp /vagrant/cni_conf/odlOvSCNI /opt/cni/bin

  - Also copy the .conf files to the /etc/cni/net.d/ directory
     - ex: vagrant@k8se2emaster:~$ sudo cp /vagrant/exmple/master.ovs-cni.conf /etc/cni/net.d/ovs-cni.conf
     - ex: vagrant@k8se2eminion1:~$ sudo cp /vagrant/exmple/worker1.ovs-cni.conf /etc/cni/net.d/ovs-cni.conf
     - ex: vagrant@k8se2eminion2:~$ sudo cp /vagrant/exmple/worker2.ovs-cni.conf /etc/cni/net.d/ovs-cni.conf

  - Specifying the K8s Node IP-address that we want to be used by k8s cluster nodes
     - Open the /etc/systemd/system/kubelet.service.d/10-kubeadm.conf file
        - $ sudo vi /etc/systemd/system/kubelet.service.d/10-kubeadm.conf
     - Add the following line then save and close. You will need to change the IP-address based on the node
        - for k8se2emaster: Environment="KUBELET_EXTRA_ARGS=--fail-swap-on=false --node-ip=192.168.30.11"
        - for k8se2eminion1: Environment="KUBELET_EXTRA_ARGS=--fail-swap-on=false --node-ip=192.168.30.12"
        - for k8se2eminion2: Environment="KUBELET_EXTRA_ARGS=--fail-swap-on=false --node-ip=192.168.30.13"
     - Restart kubelet service
        - $ sudo systemctl daemon-reload
        - $ sudo systemctl restart kubelet
     - Repeat the steps for all nodes.

  - Start the Kubernetes cluster using the following command
     - vagrant@k8se2emaster:~$ sudo kubeadm init --apiserver-advertise-address=192.168.30.11
        - Note: if you get the swap issue do the following
          - disable swap and restart kubelet
            - $ sudo swapoff -a
            - $ sudo systemctl daemon-reload
            - $ sudo systemctl restart kubelet
          - create k8s cluster using (version should be change based on your deployment)
            - $ sudo kubeadm init --apiserver-advertise-address=192.168.30.11 --ignore-preflight-errors Swap
        - Note: read the command output in order to use the kubectl command after
        - Note: in the minion VMs you will use the join command instead ex:
          - vagrant@k8se2eminion2:~$ sudo kubeadm join --token {given_token} 192.168.30.11:6443 --ignore-preflight-errors Swap

  - In order to setup the odlKubeProxy run the following commands
     - vagrant@k8se2emaster:~$ /vagrant/initkubectl.sh
     - vagrant@k8se2emaster:~$ kubectl create -f /vagrant/ovs-kube-proxy.yaml

  - Run the following command to check kube-system pods
     - vagrant@k8se2emaster:~$ kubectl -n kube-system get pods -o wide

  - In order to create pods example will do the following commands:
    - First we will label the k8s cluster nodes to create pods into different nodes
        - vagrant@k8se2emaster:~$sudo kubectl label nodes k8se2eminion1 ndName=minion1
        - vagrant@k8se2emaster:~$sudo kubectl label nodes k8se2eminion2 ndName=minion2

    - Will create pods using the yaml files exist under ~/example directory
         - vagrant@k8se2emaster:~$ kubectl create -f example/apache-pod.yaml
         - vagrant@k8se2emaster:~$ kubectl create -f example/curl-pod.yaml
      
    - Check the pods status by executing
        - vagrant@k8se2emaster:~$ sudo kubectl get pods -o wide
        - you should see that apache-pod created at node minio1 and curl-pod created at node minio2
      
    - Create apache service using the following command
        - vagrant@k8se2emaster:~$ kubectl create -f example/apache-e-w.yaml
      
    - Check the service info by executing
         - vagrant@k8se2emaster:~$ kubectl get services -o wide
         - you should the see the apacheservice IP-address and external IP-address.
       
    - Now we will test the end to end communication.
       - Execute the following command to check the communication between curl-pod and apache-services. You should get "apache" response 
         - vagrant@k8se2emaster:~$ kubectl exec -it curpod -- curl http://{service-ip}:8800
       - Execute the following command from your host machine. You should get "apache" response
         - ubuntu@ubuntu:~$ curl http://192.168.40.12:8800
         
    - If you are interested to check ovs bridges & flow rules at k8s nodes. Execute the following commands on the k8s minion nodes.
       - vagrant@k8se2eminion1:~$ sudo ovs-vsctl show
         - you should see the pods ports attached to br-int
       - vagrant@k8se2eminion1:~$ sudo ovs-ofctl dump-flows br-int
         - you should see the flow rules that symkubeproxy push to manage pod communications
         - note: check the flow rules before and after testing the service communications step (6)
       - vagrant@k8se2eminion1:~$ sudo ovs-ofctl dump-flows br-ext