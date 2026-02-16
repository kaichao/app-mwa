IMAGE_NAME:=app-mwa/router-beam-form-app-base

build:
	DOCKER_BUILDKIT=1 docker build --network=host \
		--build-arg PREFIX= --build-arg ARCH=amd64 \
		-f Dockerfile -t $(IMAGE_NAME) ../..

dist: build
	docker save $(IMAGE_NAME) | zstdmt | pv | ssh p419 'zstd -d | docker load'
	date

push:
	docker push $(IMAGE_NAME)

clean:
	docker rmi $(IMAGE_NAME)
