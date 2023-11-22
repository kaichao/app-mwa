IMAGE_NAME:=app-mwa/message-router:2309

build:
	DOCKER_BUILDKIT=1 docker build --network=host -t $(IMAGE_NAME) .

push:
	docker push $(IMAGE_NAME)

clean:
	docker rmi $(IMAGE_NAME)
