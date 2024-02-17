IMAGE_NAME:=app-mwa/down-sampler:latest

build:
	docker build --network=host -t $(IMAGE_NAME) .

build-singularity:
	mkdir -p ~/singularity/app-mwa/
	rm -f ~/singularity/app-mwa/down-sampler.sif
	date
	singularity build ~/singularity/app-mwa/down-sampler.sif docker-daemon://$(IMAGE_NAME)
	date

	mkdir -p /raid0/singularity/app-mwa/
	mv -f ~/singularity/app-mwa/down-sampler.sif /raid0/singularity/app-mwa/

dist:
	@# docker save $(IMAGE_NAME) | zstdmt | pv | ssh c0 'zstd -d | docker load'
	@ date
	docker save $(IMAGE_NAME) | zstd > /raid0/tmp/my.img.zst
	@ date
	ssh node1 'zstd -dc /raid0/tmp/my.img.zst | docker load'
	ssh node2 'zstd -dc /raid0/tmp/my.img.zst | docker load'
	ssh node3 'zstd -dc /raid0/tmp/my.img.zst | docker load'
	@ date

push:
	docker push $(IMAGE_NAME)

clean:
	docker rmi $(IMAGE_NAME)

all: build dist

