=========================
Setting Up COE in Vagrant
=========================

#. Run vagrant up
    - If Fedora wants to use libvirtd as the default provider then set it explicitly with --provider=virtualbox or export VAGRANT_DEFAULT_PROVIDER=virtualbox.

#. ssh to the VM as vagrant ssh k8s-master
#. for each VM, run the following commands:
    - start the Kubernetes cluster using the following command::

       vagrant@k8sMaster:~$ sudo kubeadm init --apiserver-advertise-address=192.168.33.11

      Note: If you get the swap issue, then do the following:

      #. open the /etc/systemd/system/kubelet.service.d/10-kubeadm.conf file::

            $ sudo vi /etc/systemd/system/kubelet.service.d/10-kubeadm.conf

      #. add the following line then save and close::

            Environment="KUBELET_EXTRA_ARGS=--fail-swap-on=false"

      #. disable swap and restart kubelet::

            $ sudo swapoff -a
            $ sudo systemctl daemon-reload
            $ sudo systemctl restart kubelet

      #. create k8s cluster using::

            $ sudo kubeadm init --apiserver-advertise-address=192.168.33.11 --ignore-preflight-errors Swap

         Note: Read the command output in order to use the kubectl command after.

         Note: In the minion VMs, you will use the join command instead ex::

          vagrant@k8sMinion2:~$ sudo kubeadm join --token {given_token} 192.168.33.11:6443

    - In order to create pods example use the following commands::

        vagrant@k8sMaster:~$ sudo kubectl create namespace sock-shop

        vagrant@k8sMaster:~$ sudo kubectl apply -n sock-shop -f "https://github.com/microservices-demo/microservices-demo/blob/master/deploy/kubernetes/complete-demo.yaml?raw=true"

    - Check the pods status by executing::

        vagrant@k8sMaster:~$ sudo kubectl -n sock-shop get pods -o wide


Verification and Troubleshooting
--------------------------------

#. Tunnels: a full overlay mesh is established between all three nodes.
   Each node will have two ports beginning with tun on br-int.::

    $ vagrant ssh k8s-master
    $ sudo ovs-vsctl show
    [vagrant@k8sMaster cni]$ sudo ovs-vsctl show
    ba282931-cf8e-4440-8982-4ae3ea71e014
        Manager "tcp:192.168.33.11:6640"
        Bridge br-int
            Controller "tcp:192.168.33.11:6653"
            fail_mode: secure
            Port "tun27b98be53d8"
                Interface "tun27b98be53d8"
                    type: vxlan
                    options: {key=flow, local_ip="192.168.33.11", remote_ip="192.168.33.13"}
            Port "veth39657fe3"
                Interface "veth39657fe3"
                    error: "could not open network device veth39657fe3 (No such device)"
            Port "tun75d56d08394"
                Interface "tun75d56d08394"
                    type: vxlan
                    options: {key=flow, local_ip="192.168.33.11", remote_ip="192.168.33.12"}
            Port br-int
                Interface br-int
                    type: internal
        ovs_version: "2.8.2"

#. veth ports: each node will have a veth on br-int.
    - Use the same command for the tunnels verification step.

#. nodes

    - kubectl get nodes::

       [vagrant@k8sMaster cni]$ kubectl get nodes
       NAME         STATUS     ROLES     AGE       VERSION
       k8smaster    Ready      master    3d        v1.9.4
       k8sminion1   NotReady   <none>    3d        v1.9.4
       k8sminion2   NotReady   <none>    3d        v1.9.4
