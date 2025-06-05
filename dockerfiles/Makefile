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
	for i in {1..4};do echo i=$$i; docker-dist node$$i hub.cstcloud.cn/scalebox/file-copy;done
	make -C mwa-vcstools dist-dcu
	make -C down-sampler dist-dcu

dist-singularity-p419:
	singularity build -F ~/singularity/scalebox/file-copy.sif  docker-daemon://hub.cstcloud.cn/scalebox/file-copy:latest
	# singularity build -F ~/singularity/scalebox/node-agent.sif docker-daemon://hub.cstcloud.cn/scalebox/node-agent:latest
	scp  ~/singularity/scalebox/file-copy.sif login1:singularity/scalebox/
	# scp  ~/singularity/scalebox/node-agent.sif login1:singularity/scalebox/
	make -C mwa-vcstools dist-singularity-p419
	make -C down-sampler dist-singularity-p419

build-all-images:
	# make -C ~/WORK/workspace/scalebox/dockerfiles/files/dir-list
	# docker-dist p419 hub.cstcloud.cn/scalebox/dir-listx
	make -C ~/WORK/workspace/scalebox/dockerfiles/files/file-copy
	docker-dist p419 hub.cstcloud.cn/scalebox/file-copy
	# make -C ~/WORK/workspace/scalebox-private/apps/cluster-admin/cluster-head dist
	# make -C ~/WORK/workspace/scalebox-private/apps/cluster-admin/node-agent dist
	make -C mwa-vcstools dist-p419
	make -C down-sampler dist-p419

# In MacOS
dist-p419: build-all-images
	ssh p419-scalebox 'make -C ~/app-mwa/dockerfiles dist-singularity-images'

build-singularity-images:
	singularity build -F ~/singularity/scalebox/file-copy.sif docker-daemon://hub.cstcloud.cn/scalebox/file-copy:latest
	make -C 

dist-p419-h12:
	make -C ~/scalebox/dockerfiles/files/file-copy
	docker-dist p419 hub.cstcloud.cn/scalebox/file-copy
	make -C mwa-vcstools dist-p419
	make -C down-sampler dist-p419
	ssh p419 'make -C ~/app-mwa/dockerfiles dist-singularity-p419'
