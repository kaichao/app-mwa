app-create:
	make -C ~/scalebox/server clean-agents
	make -C message-router-go
	scalebox app create -e dcu.env

clean-main:
	cd /raid0/scalebox/mydata/ && \
	sudo rm -rf mwa/*
