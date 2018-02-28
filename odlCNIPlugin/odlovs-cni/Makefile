default: build

build: build-image
	docker run --name tmp-odlovs-cni odlovs-cni /bin/true
	docker cp tmp-odlovs-cni:/go/bin/odlovs-cni odlovs-cni
	docker rm tmp-odlovs-cni


build-image:
	docker build -t odlovs-cni .
