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
	make -C mwa-vcstools dist-dcu
	make -C down-sampler dist-dcu
	make -C ~/scalebox/dockerfiles/files/file-copy dist-dcu
