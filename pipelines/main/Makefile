app-create:
	make -C ~/scalebox/server clean-agents
	make -C message-router-go

	ssh node1 rm -rf /tmp/scalebox/mydata/mwa/dat/ /dev/shm/scalebox/mydata/mwa/{1ch,1chx,24ch}/
	ssh node2 rm -rf /tmp/scalebox/mydata/mwa/dat/ /dev/shm/scalebox/mydata/mwa/{1ch,1chx,24ch}/
#	scalebox app create -e dcu-local-mode.env

	rm -rf /raid0/scalebox/mydata/mwa/{1ch,1chx,24ch,dat}/
	scalebox app create -e dcu.env

clean-main:
	cd /raid0/scalebox/mydata/ && \
	sudo rm -rf mwa/*
