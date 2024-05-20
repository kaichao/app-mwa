IMAGE_NAME:=app-mwa/down-sampler:latest
# set default dir
SINGULARITY_DIR ?= /raid0/root/singularity

build:
	docker build --network=host -t $(IMAGE_NAME) .

singularity:
	mkdir -p ~/singularity/app-mwa/
	rm -f ~/singularity/app-mwa/down-sampler.sif
	date
	singularity build ~/singularity/app-mwa/down-sampler.sif docker-daemon://$(IMAGE_NAME)
	date

	# mkdir -p $(SINGULARITY_DIR)/app-mwa/
	# mv -f ~/singularity/app-mwa/down-sampler.sif $(SINGULARITY_DIR)/app-mwa/
	ssh login1 mkdir -p singularity/app-mwa/
	scp  ~/singularity/app-mwa/down-sampler.sif login1:singularity/app-mwa/

dist: build
	@ docker save $(IMAGE_NAME) | zstdmt | pv | ssh c0 'zstd -d | docker load'
	@ date
	# docker save $(IMAGE_NAME) | zstd -T4 > /raid0/tmp/my.img.zst
	# ssh node1 'zstd -dc /raid0/tmp/my.img.zst | docker load'
	# ssh node2 'zstd -dc /raid0/tmp/my.img.zst | docker load'
	# ssh node3 'zstd -dc /raid0/tmp/my.img.zst | docker load'
	#ssh node4 'zstd -dc /raid0/tmp/my.img.zst | docker load'
	@ date

push:
	docker push $(IMAGE_NAME)

clean:
	docker rmi $(IMAGE_NAME)

all: build dist

