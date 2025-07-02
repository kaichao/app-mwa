IMAGE_NAME:=app-mwa/message-router-beam-form-app

build:
	DOCKER_BUILDKIT=1 docker build --network=host \
		--build-arg PREFIX= --build-arg ARCH=amd64 \
		-f Dockerfile -t $(IMAGE_NAME) ../..

build-inner:
	http_proxy=http://10.100.1.30:8118 \
	https_proxy=http://10.100.1.30:8118 \
	DOCKER_BUILDKIT=1 \
	docker build --network=host \
		--build-arg PREFIX= \
		--build-arg ARCH=amd64 \
		--build-arg GOPROXY=direct \
	    --add-host github.com:140.82.116.3 \
	    --add-host proxy.golang.org:142.250.189.241 \
    	--add-host deb.debian.org:146.75.46.132 \
    	--add-host goproxy.io:104.26.4.191 \
		-f Dockerfile -t $(IMAGE_NAME) ../..

dist: build
	docker save $(IMAGE_NAME) | zstdmt | pv | ssh p419 'zstd -d | docker load'
	date

dist-dcu: build
	docker save $(IMAGE_NAME) | zstdmt | pv | ssh dcu 'zstd -d | docker load'
	date

push:
	docker push $(IMAGE_NAME)

clean:
	docker rmi $(IMAGE_NAME)
