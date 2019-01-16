default: build

build: binary container

binary:
	dep ensure -vendor-only
	go build -o bin/odlovs-cni

container:
	docker build . -t  odlovs-cni

# The directory /home/vagrant/.kube  may need to be change based on the env
run: build
	sudo docker run --network=host -it  -v /home/vagrant/.kube:/root/.kube/ -v /etc/cni/net.d/:/etc/cni/net.d/ -v /opt/cni/bin/:/opt/cni/bin/ odlovs-cni
