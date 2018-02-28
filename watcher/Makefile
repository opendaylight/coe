default: build

build: build-image
	docker run --name tmp-watcher watcher /bin/true
	docker cp tmp-watcher:/go/bin/watcher watcher
	docker rm tmp-watcher


build-image:
	docker build -t watcher .

run: build-image
	docker run --network=host -v ${HOME}/.kube:/root/.kube -v ${HOME}/.minikube:${HOME}/.minikube watcher watcher odl
