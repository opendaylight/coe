default: build

build:
	docker build -t watcher .

run: build
	docker run --network=host -v ${HOME}/.kube:/.kube -v ${HOME}/.minikube:${HOME}/.minikube watcher odl
