IMAGE_NAME:=app-mwa/repack

build:
	docker build --network=host -t $(IMAGE_NAME) .
dist:
	docker save $(IMAGE_NAME) > /raid0/tmp/repack.tar
	for i in {01..15}; do echo $$i; ssh r$$i docker load < /raid0/tmp/repack.tar; done
	rm -f $(IMAGE_FILE)
push:
	docker push $(IMAGE_NAME)
clean:
	docker rmi $(IMAGE_NAME)
