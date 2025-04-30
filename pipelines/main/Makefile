du:
	@ for i in 1 3 4;do \
		echo NODE$$i; \
		ssh node$$i du -ms /dev/shm/scalebox/mydata/mwa/{1ch,1chx,24ch} /tmp/scalebox/mydata/mwa/dat/1257010784/*; \
	done

app-create: clean
	# make -C ~/scalebox/server clean-agents
	make -C message-router-go
	# scalebox app create -e dcu-singularity.env
	scalebox app create -e dcu.env

clean:
	@ # local-mode
	ssh node1 rm -rf /tmp/scalebox/mydata/mwa/dat/ /dev/shm/scalebox/mydata/mwa/{1ch,1chx,24ch}/
	# ssh node2 rm -rf /tmp/scalebox/mydata/mwa/dat/ /dev/shm/scalebox/mydata/mwa/{1ch,1chx,24ch,dedisp,png}/
	ssh node3 rm -rf /tmp/scalebox/mydata/mwa/dat/ /dev/shm/scalebox/mydata/mwa/{1ch,1chx,24ch}/
	ssh node4 rm -rf /tmp/scalebox/mydata/mwa/dat/ /dev/shm/scalebox/mydata/mwa/{1ch,1chx,24ch}/
	@ # non-local-mode
	# rm -rf /raid0/scalebox/mydata/mwa/{1ch,1chx,24ch}/

dist-singularity-images:
	singularity build -F ~/singularity/scalebox/file-copy.sif  docker-daemon://hub.cstcloud.cn/scalebox/file-copy:latest
	singularity build -F ~/singularity/scalebox/node-agent.sif docker-daemon://hub.cstcloud.cn/scalebox/node-agent:latest
	scp  ~/singularity/scalebox/file-copy.sif login1:singularity/scalebox/
	scp  ~/singularity/scalebox/node-agent.sif login1:singularity/scalebox/
	make -C ~/app-mwa/dockerfiles/mwa-vcstools dist-singularity
	make -C ~/app-mwa/dockerfiles/down-sampler dist-singularity
