app-create:
	# scalebox app create
	# scalebox app create -e dcu.env
	scalebox app create -e p419.env

clean:
	# rm -rf /raid0/scalebox/mydata/mwa/{1ch,1chx,24ch}/
	ssh login1 rm -rf /work1/cstu0036/mydata/mwa/{1ch,1chx,24ch}/

dataset-import:
	CLUSTER=dcu scalebox app create dataset.yaml
