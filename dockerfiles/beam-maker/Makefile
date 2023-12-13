IMAGE_NAME:=app-mwa/beam-maker

build:
	docker build --network=host -t $(IMAGE_NAME) .

dist:
	docker save $(IMAGE_NAME) | pv | ssh node1 'docker load'
	docker save $(IMAGE_NAME) | pv | ssh node2 'docker load'


push:
	docker push $(IMAGE_NAME)

clean:
	docker rmi $(IMAGE_NAME)
