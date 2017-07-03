#!/bin/bash

PUBLIC_IP=$1
PUBLIC_SUBNET_MASK=$2
ODL_IP=$3

cat > setup_k8s_master_args.sh <<EOL
PUBLIC_IP=$1
PUBLIC_SUBNET_MASK=$2
ODL_IP=$3
EOL


# Install etcd cluster
sudo docker run --net=host -d gcr.io/google_containers/etcd:2.0.12 /usr/local/bin/etcd \
                --addr=127.0.0.1:4001 --bind-addr=0.0.0.0:4001 --data-dir=/var/etcd/data

# Start k8s daemons
pushd k8s/server/kubernetes/server/bin
echo "Starting kube-apiserver ..."
nohup sudo ./kube-apiserver --service-cluster-ip-range=192.168.200.0/24 \
                            --address=0.0.0.0 --etcd-servers=http://127.0.0.1:4001 \
                            --v=2 2>&1 0<&- &>/dev/null &
sleep 5

echo "Starting kube-controller-manager ..."
nohup sudo ./kube-controller-manager --master=127.0.0.1:8080 --v=2 2>&1 0<&- &>/dev/null &
sleep 5

echo "Starting kube-scheduler ..."
nohup sudo ./kube-scheduler --master=127.0.0.1:8080 --v=2 2>&1 0<&- &>/dev/null &
sleep 5

# set the ovs instance in passive mode for now.
sudo ovs-vsctl set-manager tcp:$ODL_IP:6640

# Create a OVS physical bridge and move IP address of enp0s9 to br-enp0s9
echo "Creating physical bridge ..."
sudo ovs-vsctl add-br br-enp0s9
sudo ovs-vsctl add-port br-enp0s9 enp0s9
sudo ip addr flush dev enp0s9
sudo ifconfig br-enp0s9 $PUBLIC_IP netmask $PUBLIC_SUBNET_MASK up
sudo ovs-vsctl set-controller br-enp0s9 tcp:$ODL_IP:6653
sleep 5
popd

mkdir examples
# Setup some example yaml files
cat << APACHEPOD >> ~/examples/apache-pod.yaml
apiVersion: v1
kind: Pod
metadata:
  name: apachetwin
  labels:
    name: webserver
spec:
  containers:
  - name: apachetwin
    image: fedora/apache
APACHEPOD

cat << NGINXPOD >> ~/examples/nginx-pod.yaml
apiVersion: v1
kind: Pod
metadata:
  name: nginxtwin
  labels:
    name: webserver
spec:
  containers:
  - name: nginxtwin
    image: nginx
NGINXPOD

cat << APACHEEW >> ~/examples/apache-e-w.yaml
apiVersion: v1
kind: Service
metadata:
  labels:
    name: apacheservice
    role: service
  name: apacheservice
spec:
  ports:
    - port: 8800
      targetPort: 80
      protocol: TCP
      name: tcp
  selector:
    name: webserver
APACHEEW

cat << APACHENS >> ~/examples/apache-n-s.yaml
apiVersion: v1
kind: Service
metadata:
  labels:
    name: apacheexternal
    role: service
  name: apacheexternal
spec:
  ports:
    - port: 8800
      targetPort: 80
      protocol: TCP
      name: tcp
  selector:
    name: webserver
  type: NodePort
APACHENS