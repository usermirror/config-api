NAME=config-api
DEPLOYMENT=deployment/$(NAME)-deployment
GOCMD=go
GET_HASH=$(GOCMD) run $(GOPATH)/src/github.com/segmentio/ksuid/cmd/ksuid/main.go
HASH:=$(shell $(GET_HASH))
GCP_PROJECT=um-west1-prod
GCR_REPO=us.gcr.io/$(GCP_PROJECT)/$(NAME)

gcr-build:
	@echo "🛠 $(HASH)"
	@docker build -q -t $(GCR_REPO):$(HASH) -t $(GCR_REPO):latest . > /dev/null
	@echo "✅ $(GCR_REPO):$(HASH)"

gcr-push:
	@echo "🛫 GCR · $(HASH)"
	@gcloud docker -- push $(GCR_REPO) > /dev/null
	@echo "🛬 GCR · $(GCR_REPO):$(HASH)"

gcr: gcr-build gcr-push

gcr-rolling-update:
	kubectl set image $(DEPLOYMENT) $(NAME)=$(GCR_REPO):$(HASH)

gcr-deploy: gcr gcr-rolling-update

rollout-status:
	kubectl rollout status $(DEPLOYMENT)

rollout-history:
	kubectl rollout history $(DEPLOYMENT)
 
deps:
	go get -u github.com/codegangsta/gin
	go get .

watch:
	gin -p 4200 -a 8888 run cmd/config/main.go 