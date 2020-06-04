multi:
	# docker buildx create --name mbuilder
	docker buildx use mbuilder
	docker buildx build  --platform linux/amd64,linux/arm64 -t ${TAG} --push .