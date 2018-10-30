#!/bin/bash

# Some example yaml files
mkdir examples
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
    imagePullPolicy: IfNotPresent
  nodeSelector:
    ndName: minion1
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
  externalIPs:
  - 192.168.40.12
  - 192.168.40.13
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

cat << CURLPOD >> ~/examples/curl-pod.yaml
apiVersion: v1
kind: Pod
metadata:
  name: curlpod
spec:
  containers:
  - image: radial/busyboxplus:curl
    command:
      - sleep
      - "3600"
    imagePullPolicy: IfNotPresent
    name: curlcontainer
  restartPolicy: Always
  nodeSelector:
    ndName: minion2
CURLPOD

cat << BUSYBOX >> ~/examples/busybox.yaml
apiVersion: v1
kind: Pod
metadata:
  name: busybox
  namespace: default
spec:
  containers:
  - image: busybox
    command:
      - sleep
      - "3600"
    imagePullPolicy: IfNotPresent
    name: busybox
  restartPolicy: Always
BUSYBOX