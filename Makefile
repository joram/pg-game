run:
	docker build -t veilstream/psql-text-based-adventure:latest .
	docker run -it --rm \
		-v $(shell pwd):/app/ \
		-p 2850:5432 \
		veilstream/psql-text-based-adventure:latest

connect:
	psql -h localhost -p 2850 -U postgres -d postgres