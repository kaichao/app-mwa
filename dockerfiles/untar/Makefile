IMAGE_NAME:=hub.cstcloud.cn/scalebox/untar
IMAGE_FILE:=/gfsdata/tmp/my.img
build:
	docker build --network=host -t $(IMAGE_NAME) .
dist:
	docker save $(IMAGE_NAME) > /raid0/tmp/untar.tar
	ssh r01 docker load -i /raid0/tmp/untar.tar
	ssh r02 docker load -i /raid0/tmp/untar.tar
	ssh r03 docker load -i /raid0/tmp/untar.tar
	ssh r04 docker load -i /raid0/tmp/untar.tar
	ssh r05 docker load -i /raid0/tmp/untar.tar
	ssh r06 docker load -i /raid0/tmp/untar.tar
	ssh r07 docker load -i /raid0/tmp/untar.tar
	ssh r08 docker load -i /raid0/tmp/untar.tar
	ssh r09 docker load -i /raid0/tmp/untar.tar
	ssh r10 docker load -i /raid0/tmp/untar.tar
	ssh r11 docker load -i /raid0/tmp/untar.tar
	ssh r12 docker load -i /raid0/tmp/untar.tar
	#ssh r13 docker load -i /raid0/tmp/unpack.tar
	#ssh r14 docker load -i /raid0/tmp/unpack.tar
	#ssh r15 docker load -i /raid0/tmp/unpack.tar
push:
	docker push $(IMAGE_NAME)
clean:
	docker rmi $(IMAGE_NAME)
