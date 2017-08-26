default: build

build:
	docker build -t watcher .

run: build
	docker run --network=host -v ${HOME}/.kube:/root/.kube -v ${HOME}/.minikube:${HOME}/.minikube watcher watcher odl
