dirs:=mwa-vcstools copy-unpack down-sampler presto-search repack unpack

build:
	@for dir in $(dirs); do \
		$(MAKE) -C $$dir build; \
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
