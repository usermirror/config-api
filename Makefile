NAME=config-api
DEPLOYMENT=deployment/$(NAME)
GOCMD=go
GET_HASH=git rev-parse HEAD
HASH:=$(shell $(GET_HASH))
GCP_PROJECT=um-west1-prod
GCR_REPO=us.gcr.io/$(GCP_PROJECT)/$(NAME)
DOCKERHUB_ORG=usermirror
DOCKERHUB_IMAGE=$(DOCKERHUB_ORG)/$(NAME)

build:
	@make docker-build docker-push

docker-build:
	@echo "Building... ($(HASH))"
	@docker build -t $(DOCKERHUB_IMAGE):$(HASH) -t $(DOCKERHUB_IMAGE):latest  .
	@docker tag $(DOCKERHUB_IMAGE):$(HASH) $(GCR_REPO):$(HASH)
	@echo "Built complete ($(HASH))"

docker-push:
	@echo "🛫 Docker Hub · $(HASH)"
	@docker push $(DOCKERHUB_IMAGE):$(HASH)
	@echo "🛬 Docker Hub · $(DOCKERHUB_IMAGE):$(HASH)"

gcr-push:
	@echo "🛫 GCR · $(HASH)"
	@gcloud docker -- push $(GCR_REPO) > /dev/null
	@echo "🛬 GCR · $(GCR_REPO):$(HASH)"

gcr-push-latest:
	@echo "🛫 GCR · latest"
	@gcloud docker -- push $(GCR_REPO):latest > /dev/null
	@echo "🛬 GCR · $(GCR_REPO):latest"

rollout-status:
	kubectl rollout status $(DEPLOYMENT)

rollout-history:
	kubectl rollout history $(DEPLOYMENT)
 
deps:
	go get -u github.com/codegangsta/gin
	go get .

watch:
	gin -p 4200 -a 8888 run cmd/config/main.go 
