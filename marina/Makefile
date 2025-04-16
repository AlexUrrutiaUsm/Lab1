SERVICE_NAME = marina
PROTO_FILE = proto/interaccion.proto
DOCKER_IMAGE = $(SERVICE_NAME)-image
DOCKER_CONTAINER = $(SERVICE_NAME)-container
PORT = 50051

docker-marina: docker-clean docker-build docker-run

docker-build:
        docker build -t $(DOCKER_IMAGE) .

docker-run:
        docker run -d -p $(PORT):$(PORT) --name $(DOCKER_CONTAINER) $(DOCKER_IMAGE)

docker-clean:
        docker stop $(DOCKER_CONTAINER) || true
        docker rm $(DOCKER_CONTAINER) || true
