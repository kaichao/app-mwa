IMAGE_NAME:=app-mwa/down-sampler:latest
# set default dir
SINGULARITY_DIR ?= /raid0/root/singularity

build:
	docker build --network=host -t $(IMAGE_NAME) .

dist-p419: build
	@ docker save $(IMAGE_NAME) | zstdmt | pv | ssh p419 'zstd -d | docker load'

dist-dcu: build
	@ docker save $(IMAGE_NAME) | zstdmt | pv | ssh dcu 'zstd -d | docker load'

dist-dcu-inner:
	@ date
	docker save $(IMAGE_NAME) | zstd -T4 > /raid0/tmp/my.img.zst
	pdsh -w node[1-4] -l root 'zstd -dc /raid0/tmp/my.img.zst | docker load' | dshbak -c
	@ date

build-singularity:
	mkdir -p ~/singularity/app-mwa/
	@date
	singularity build -F ~/singularity/app-mwa/down-sampler.sif docker-daemon://$(IMAGE_NAME)
	@date

dist-singularity-p419: build-singularity
	ssh login1 mkdir -p singularity/app-mwa/
	scp  ~/singularity/app-mwa/down-sampler.sif login1:singularity/app-mwa/

dist-singularity-dcu: build-singularity
	mkdir -p $(SINGULARITY_DIR)/app-mwa/
	mv  ~/singularity/app-mwa/down-sampler.sif $(SINGULARITY_DIR)/app-mwa/

clean:
	docker rmi $(IMAGE_NAME)
