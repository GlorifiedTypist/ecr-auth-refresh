TAG = latest
BIN = ecr-auth-refresh

.PHONY: all
all: container tag push

.PHONY: build
build: main.go
	go build -o $(BIN) -v .

.PHONY: container
container: build
	docker build -t $(BIN) .

.PHONY: tag
tag:
	docker tag $(BIN):latest jeffvader/$(BIN):latest

.PHONY: push
push:
	docker push jeffvader/$(BIN):latest

.PHONY: test-deploy
test-deploy:
	kubectl delete -f ./deployment/deployment.yaml; kubectl apply -f ./deployment/deployment.yaml

.PHONY: clean
clean:
	rm -f $(BIN)

