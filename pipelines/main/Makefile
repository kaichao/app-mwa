app-create:
	make -C ~/scalebox/server clean-agents
	rm -rf ~/.scalebox/var/*
	ssh node1 rm -rf /tmp/scalebox/ /dev/shm/scalebox/
	ssh node2 rm -rf /tmp/scalebox/ /dev/shm/scalebox/
	
	make -C message-router-go
	scalebox app create -e dcu-local-mode.env

clean-main:
	cd /raid0/scalebox/mydata/ && \
	sudo rm -rf mwa/*
