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
	# make -C $(SCALEBOX_DIR)/scalebox/dockerfiles/files/dir-list
	# docker-dist p419 hub.cstcloud.cn/scalebox/dir-listx
	# make -C $(SCALEBOX_DIR)/dockerfiles/files/file-copy
	# make -C ~/WORK/workspace/scalebox-private/apps/cluster-admin/cluster-head dist
	# make -C ~/WORK/workspace/scalebox-private/apps/cluster-admin/node-agent dist
	make -C mwa-vcstools
	make -C down-sampler

dist-dcu: build-images
	make -C $(SCALEBOX_DIR)/dockerfiles/files/file-copy
	for i in {1..4};do echo i=$$i; docker-dist node$$i hub.cstcloud.cn/scalebox/file-copy;done
	make -C mwa-vcstools dist-dcu
	make -C down-sampler dist-dcu

dist-p419: build-images
	docker-dist p419 hub.cstcloud.cn/scalebox/file-copy
	make -C mwa-vcstools dist-p419
	make -C down-sampler dist-p419

dist-singularity-p419:
	singularity build -F ~/singularity/scalebox/file-copy.sif  docker-daemon://hub.cstcloud.cn/scalebox/file-copy:latest
	scp  ~/singularity/scalebox/file-copy.sif login1:singularity/scalebox/
	# singularity build -F ~/singularity/scalebox/node-agent.sif docker-daemon://hub.cstcloud.cn/scalebox/node-agent:latest
	# scp  ~/singularity/scalebox/node-agent.sif login1:singularity/scalebox/
	make -C mwa-vcstools dist-singularity-p419
	make -C down-sampler dist-singularity-p419

# In MacOS/h12
dist-all-p419: dist-p419
	ssh p419 'make -C ~/app-mwa/dockerfiles dist-singularity-p419'

# make dist-all-p419 HOST=h12
