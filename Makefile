run:
	docker build -t veilstream/psql-text-based-adventure:latest .
	docker run -it --rm \
		-v $(shell pwd):/app/ \
		-p 2850:5432 \
		-p 8080:80 \
		veilstream/psql-text-based-adventure:latest

connect:
	psql -h localhost -p 2850 -U postgres -d postgres

# take key; use key on door; go north; go east; look; take screwdriver; go west; use screwdriver on lantern; take lantern;
connect_prod:
	psql -h psql-text-based-adventure.proxy.veilstream.com -p 5432 -U veilstream -d veilstream

deploy:
	docker build -f ./Dockerfile.prod -t veilstream/psql-text-based-adventure:latest .
	docker push veilstream/psql-text-based-adventure:latest
	aws ecs update-service --cluster veilstream-cluster --service psql-text-based-adventure --force-new-deployment --region ca-central-1
	aws ecs update-service --cluster veilstream-cluster --service dnd-game --force-new-deployment --region ca-central-1
