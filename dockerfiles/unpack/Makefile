IMAGE_NAME:=app-mwa/unpack:latest
# set default dir
SINGULARITY_DIR ?= /raid0/root/singularity

build:
	docker build --network=host -t $(IMAGE_NAME) .

singularity:
	mkdir -p ~/singularity/app-mwa/
	rm -f ~/singularity/app-mwa/unpack.sif
	date
	singularity build ~/singularity/app-mwa/unpack.sif docker-daemon://$(IMAGE_NAME)
	date

	mkdir -p $(SINGULARITY_DIR)/app-mwa/
	mv -f ~/singularity/app-mwa/unpack.sif $(SINGULARITY_DIR)/app-mwa/

dist:
	@# docker save $(IMAGE_NAME) | zstdmt | pv | ssh c0 'zstd -d | docker load'
	@ date
	docker save $(IMAGE_NAME) | zstd > /raid0/tmp/my.img.zst
	@ date
	ssh node1 'zstd -dc /raid0/tmp/my.img.zst | docker load'
	ssh node2 'zstd -dc /raid0/tmp/my.img.zst | docker load'
#	ssh node3 'zstd -dc /raid0/tmp/my.img.zst | docker load'
	ssh node4 'zstd -dc /raid0/tmp/my.img.zst | docker load'
	# pdsh -w node[1-3] -l root 'zstd -dc /raid0/tmp/my.img.zst | docker load' | dshbak -c
	@ date

push:
	docker push $(IMAGE_NAME)

clean:
	docker rmi $(IMAGE_NAME)

all: build dist

