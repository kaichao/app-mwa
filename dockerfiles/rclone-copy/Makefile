IMAGE_NAME:=hub.cstcloud.cn/scalebox/rclone-copy

build:
	docker build --network=host -t $(IMAGE_NAME) .

dist:
	docker save $(IMAGE_NAME) > /raid0/tmp/rclone.tar
	for i in {01..12}; do echo $$i; ssh r$$i docker load < /raid0/tmp/rclone.tar; done
	rm -f $(IMAGE_FILE)
