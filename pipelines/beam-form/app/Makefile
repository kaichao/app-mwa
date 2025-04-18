all: clean-dcu
	make -C message-router/
	
clean-dcu:
	for i in {1..4};do echo $$i;ssh node$$i 'docker rm -f $$(docker ps -qa)';done
	rm -rf /raid0/scalebox/mydata/mwa/24ch/* &
	for i in {1..4};do echo $$i; ssh node$$i 'rm -rf /dev/shm/scalebox/mydata';done
	for i in {1..4};do echo $$i; ssh node$$i 'rm -rf /tmp/scalebox/mydata';done
