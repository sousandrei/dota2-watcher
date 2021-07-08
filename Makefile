NAME=dota
REGISTRY_URL=gcr.io/sousandrei
VERSION=$(shell git rev-parse --short=7 HEAD)


build:
	docker build . -t ${NAME}

push:
	docker tag ${NAME} ${REGISTRY_URL}/${NAME}:${VERSION}
	docker push ${REGISTRY_URL}/${NAME}:${VERSION}

helm:
	helm upgrade --install ${NAME} ./chart --namespace ${NAME} --set image=${VERSION}

deploy: build push helm