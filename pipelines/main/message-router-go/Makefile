IMAGE_NAME:=app-mwa/message-router-go

build:
	DOCKER_BUILDKIT=1 docker build --network=host -t $(IMAGE_NAME) .

dist: build
	docker save $(IMAGE_NAME) | zstdmt | pv | ssh p419 'zstd -d | docker load'

push:
	docker push $(IMAGE_NAME)

clean:
	docker rmi $(IMAGE_NAME)
