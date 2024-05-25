IMAGE_NAME:=hub.cstcloud.cn/scalebox/agent
# set default dir
SINGULARITY_DIR ?= /raid0/root/singularity

singularity:


dist-dcu:
	@ date
	docker save $(IMAGE_NAME) | zstd -T4 > /raid0/tmp/my.img.zst
	pdsh -w node[1-4] -l root 'zstd -dc /raid0/tmp/my.img.zst | docker load' | dshbak -c
	@ date

