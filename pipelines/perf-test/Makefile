app-create: clean
	scalebox app create
	# scalebox app create -e dcu.env
	# scalebox app create -e p419.env

clean:
	@ # non-local-mode
	rm -rf /raid0/scalebox/mydata/mwa/{1ch,1chx,24ch}/

dataset-import:
	CLUSTER=dcu scalebox app create dataset.yaml
