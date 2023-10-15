IMAGE_NAME:=app-mwa/beam-maker

build:
	docker build --network=host -t $(IMAGE_NAME) .

dist:
	docker save $(IMAGE_NAME) | zstdmt | pv | ssh c0 'zstd -d | docker load'

push:
	docker push $(IMAGE_NAME)

clean:
	docker rmi $(IMAGE_NAME)
