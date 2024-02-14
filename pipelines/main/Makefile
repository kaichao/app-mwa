app-create: clean
	make -C ~/scalebox/server clean-agents
	make -C message-router-go
	scalebox app create -e dcu.env
	# scalebox app create -e dcu-non-local.env

clean:
	@ # local-mode
	ssh node1 rm -rf /tmp/scalebox/mydata/mwa/dat/ /dev/shm/scalebox/mydata/mwa/{1ch,1chx,24ch,dedisp,png}/ /dev/shm/scalebox/copy-unpack/
	ssh node2 rm -rf /tmp/scalebox/mydata/mwa/dat/ /dev/shm/scalebox/mydata/mwa/{1ch,1chx,24ch,dedisp,png}/ /dev/shm/scalebox/copy-unpack/
	@ # non-local-mode
	rm -rf /raid0/scalebox/mydata/mwa/{1ch,1chx,24ch,dat}/

dataset-import:
	CLUSTER=dcu scalebox app create dataset.yaml
