#dirs:=mwa-vcstools down-sampler presto-search
dirs:=mwa-vcstools down-sampler

# 默认 HOST 可以通过命令行传入，例如 make build HOST=h12
HOST ?= macos

ifneq ($(HOST),macos)
SCALEBOX_DIR := ~/scalebox
else
SCALEBOX_DIR := ~/WORK/workspace/scalebox
endif

build:
	@for dir in $(dirs); do \
		$(MAKE) -C $$dir build; \
	done

singularity:
	@for dir in $(dirs); do \
		$(MAKE) -C $$dir singularity; \
	done

dist:
	@for dir in $(dirs); do \
		$(MAKE) -C $$dir dist; \
	done

push:
	@for dir in $(dirs); do \
		$(MAKE) -C $$dir push; \
	done

pull:
	@for dir in $(dirs); do \
		$(MAKE) -C $$dir pull; \
	done

clean:
	@for dir in $(dirs); do \
		$(MAKE) -C $$dir clean; \
	done

build-images:
	make -C $(SCALEBOX_DIR)/dockerfiles/files/file-copy
	make -C mwa-vcstools
	make -C down-sampler
	# make -C $(SCALEBOX_DIR)/scalebox/dockerfiles/files/dir-list
	# make -C ~/WORK/workspace/scalebox-private/apps/cluster-admin/cluster-head dist
	# make -C ~/WORK/workspace/scalebox-private/apps/cluster-admin/node-agent dist

dist-dcu:
	make build-images HOST=dcu
	for i in {1..4};do echo i=$$i; docker-dist node$$i hub.cstcloud.cn/scalebox/file-copy;done
	make -C mwa-vcstools dist-dcu
	make -C down-sampler dist-dcu

dist-p419: build-images
	docker-dist p419 hub.cstcloud.cn/scalebox/file-copy
	docker-dist p419 app-mwa/mwa-vcstools:rocm-4.0.1
	docker-dist p419 app-mwa/down-sampler:latest

build-singularity-images:
	make -C $(SCALEBOX_DIR)/dockerfiles/files/file-copy build-singularity
	make -C mwa-vcstools build-singularity
	make -C down-sampler build-singularity

dist-singularity-dcu:
	make build-singularity-images HOST=dcu
	mkdir -p /raid0/root/singularity/{scalebox,app-mwa}/
	cp ~/singularity/scalebox/file-copy.sif /raid0/root/singularity/scalebox/
	cp ~/singularity/app-mwa/mwa-vcstools.sif /raid0/root/singularity/app-mwa/
	cp ~/singularity/app-mwa/down-sampler.sif /raid0/root/singularity/app-mwa/

dist-singularity-p419:
	make build-singularity-images
	ssh login1 'mkdir -p ~/singularity/{scalebox,app-mwa}/'
	scp  ~/singularity/app-mwa/mwa-vcstools.sif login1:singularity/app-mwa/
	scp  ~/singularity/app-mwa/down-sampler.sif login1:singularity/app-mwa/
	scp  ~/singularity/scalebox/file-copy.sif login1:singularity/scalebox/
	# scp  ~/singularity/scalebox/node-agent.sif login1:singularity/scalebox/

# In MacOS/h12
dist-all-p419: dist-p419
	ssh p419 'make -C ~/app-mwa/dockerfiles dist-singularity-p419 HOST=p419'

# make dist-all-p419 HOST=h12
