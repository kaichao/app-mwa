#dirs:=mwa-vcstools down-sampler presto-search
dirs:=mwa-vcstools down-sampler

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

dist-dcu:
	make -C ~/scalebox/dockerfiles/files/file-copy
	make -C mwa-vcstools dist-dcu
	make -C down-sampler dist-dcu

build-all-images:
	make -C ~/WORK/workspace/scalebox/dockerfiles/files/dir-list
	docker-dist p419 hub.cstcloud.cn/scalebox/dir-listx
	make -C ~/WORK/workspace/scalebox/dockerfiles/files/file-copy
	docker-dist p419 hub.cstcloud.cn/scalebox/file-copy
	make -C ~/WORK/workspace/scalebox-private/apps/cluster-admin/cluster-head dist
	make -C ~/WORK/workspace/scalebox-private/apps/cluster-admin/node-agent dist
	make -C mwa-vcstools dist-p419
	make -C down-sampler dist-p419

# In MacOS
dist-p419: build-all-images
	ssh p419-scalebox 'make -C ~/app-mwa/pipelines/main dist-singularity-images'
