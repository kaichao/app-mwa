IMAGE_NAME:=app-mwa/mwa-vcstools:rocm-4.0.1

build:
	docker build --network=host -t $(IMAGE_NAME) .

build-singularity:
	mkdir -p ~/singularity/app-mwa/
	rm -f ~/singularity/app-mwa/mwa-vcstools.sif
	date
	singularity build ~/singularity/app-mwa/mwa-vcstools.sif docker-daemon://$(IMAGE_NAME)
	date

	mkdir -p /raid0/singularity/app-mwa/
	mv -f ~/singularity/app-mwa/mwa-vcstools.sif /raid0/singularity/app-mwa/

dist:
	@# docker save $(IMAGE_NAME) | zstdmt | pv | ssh c0 'zstd -d | docker load'
	@ date
	docker save $(IMAGE_NAME) | zstd > /raid0/tmp/my.img.zst
	@ date
	ssh node1 'zstd -dc /raid0/tmp/my.img.zst | docker load'
	ssh node2 'zstd -dc /raid0/tmp/my.img.zst | docker load'
	ssh node3 'zstd -dc /raid0/tmp/my.img.zst | docker load'
	# pdsh -w node[123] -l root 'zstd -dc /raid0/tmp/my.img.zst | docker load' | dshbak -c
	@ date

push:
	docker push $(IMAGE_NAME)

clean:
	docker rmi $(IMAGE_NAME)
