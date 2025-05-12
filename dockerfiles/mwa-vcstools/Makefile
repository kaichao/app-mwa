IMAGE_NAME:=app-mwa/mwa-vcstools:rocm-4.0.1
# set default dir
SINGULARITY_DIR ?= /raid0/root/singularity

build:
	docker build --network=host -t $(IMAGE_NAME) .

dist-p419: build
	@ docker save $(IMAGE_NAME) | zstdmt | pv | ssh p419 'zstd -d | docker load'

dist-dcu: build
	@ date
	docker save $(IMAGE_NAME) | zstd -T4 > /raid0/tmp/my.img.zst
	pdsh -w node[1-4] -l root 'zstd -dc /raid0/tmp/my.img.zst | docker load' | dshbak -c
	# make singularity
	# mkdir -p $(SINGULARITY_DIR)/app-mwa/
	# mv -f ~/singularity/app-mwa/mwa-vcstools.sif $(SINGULARITY_DIR)/app-mwa/
	@ date

singularity:
	mkdir -p ~/singularity/app-mwa/
	date
	singularity build -F ~/singularity/app-mwa/mwa-vcstools.sif docker-daemon://$(IMAGE_NAME)
	date

dist-singularity: singularity
	ssh login1 mkdir -p singularity/app-mwa/
	scp  ~/singularity/app-mwa/mwa-vcstools.sif login1:singularity/app-mwa/

push:
	docker push $(IMAGE_NAME)

clean:
	docker rmi $(IMAGE_NAME)
