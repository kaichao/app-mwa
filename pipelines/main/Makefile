app-create: clean
	make -C ~/scalebox/server clean-agents
	make -C message-router-go
	scalebox app create -e dcu-local-mode.env
	# scalebox app create -e dcu.env

clean:
	@ # local-mode
	ssh node1 rm -rf /tmp/scalebox/mydata/mwa/dat/ /dev/shm/scalebox/mydata/mwa/{1ch,1chx,24ch}/ /dev/shm/scalebox/copy-unpack/
	ssh node2 rm -rf /tmp/scalebox/mydata/mwa/dat/ /dev/shm/scalebox/mydata/mwa/{1ch,1chx,24ch}/ /dev/shm/scalebox/copy-unpack/
	@ # non-local-mode
	rm -rf /raid0/scalebox/mydata/mwa/{1ch,1chx,24ch,dat}/
