app-create: clean
	make -C ~/scalebox/server clean-agents
	make -C message-router-go
	scalebox app create -e dcu-singularity.env
	# scalebox app create -e dcu.env
	# scalebox app create -e dcu-non-local.env

clean:
	@ # local-mode
	ssh node1 rm -rf /tmp/scalebox/mydata/mwa/dat/ /dev/shm/scalebox/mydata/mwa/{tar,1ch,1chx,24ch,dedisp,png}/
	ssh node2 rm -rf /tmp/scalebox/mydata/mwa/dat/ /dev/shm/scalebox/mydata/mwa/{tar,1ch,1chx,24ch,dedisp,png}/
	# ssh node3 rm -rf /tmp/scalebox/mydata/mwa/dat/ /dev/shm/scalebox/mydata/mwa/{tar,1ch,1chx,24ch,dedisp,png}/
	ssh node4 rm -rf /tmp/scalebox/mydata/mwa/dat/ /dev/shm/scalebox/mydata/mwa/{tar,1ch,1chx,24ch,dedisp,png}/
	@ # non-local-mode
	rm -rf /raid0/scalebox/mydata/mwa/{1ch,1chx,24ch,dat}/

dataset-import:
	CLUSTER=dcu scalebox app create dataset.yaml
