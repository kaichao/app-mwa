IMAGE_NAME:=app-mwa/presto-search

build:
	docker build --network=host -t $(IMAGE_NAME) .

dist:
	docker save $(IMAGE_NAME) | pv | ssh zhzhang@node1 'docker load'
	docker save $(IMAGE_NAME) | pv | ssh zhzhang@node2 'docker load'


push:
	docker push $(IMAGE_NAME)

clean:
	docker rmi $(IMAGE_NAME)
