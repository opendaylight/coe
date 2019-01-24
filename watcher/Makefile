default: build

build:
	docker build -t odlwatcher .

run: build
	docker run --network=host -v ${HOME}/.kube:/.kube -v ${HOME}/.minikube:${HOME}/.minikube odlwatcher
