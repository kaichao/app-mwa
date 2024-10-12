IMAGE_NAME:=hub.cstcloud.cn/scalebox/repack

build:
	docker build --network=host -t $(IMAGE_NAME) .
dist:
	docker save $(IMAGE_NAME) > /raid0/tmp/repack.tar
	ssh r01 docker load -i /raid0/tmp/repack.tar
	ssh r02 docker load -i /raid0/tmp/repack.tar
	ssh r03 docker load -i /raid0/tmp/repack.tar
	ssh r04 docker load -i /raid0/tmp/repack.tar
	ssh r05 docker load -i /raid0/tmp/repack.tar
	ssh r06 docker load -i /raid0/tmp/repack.tar
	ssh r07 docker load -i /raid0/tmp/repack.tar
	ssh r08 docker load -i /raid0/tmp/repack.tar
	ssh r09 docker load -i /raid0/tmp/repack.tar
	ssh r10 docker load -i /raid0/tmp/repack.tar
	ssh r11 docker load -i /raid0/tmp/repack.tar
	ssh r12 docker load -i /raid0/tmp/repack.tar
	#ssh r13 docker load -i /raid0/tmp/repack.tar
	#ssh r14 docker load -i /raid0/tmp/repack.tar
	#ssh r15 docker load -i /raid0/tmp/repack.tar
push:
	docker push $(IMAGE_NAME)
clean:
	docker rmi $(IMAGE_NAME)
