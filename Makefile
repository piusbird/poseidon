build: sign
	docker build --no-cache . -t deepblue
sign:
	./sign.sh
up: sign build
	docker run --net customnetwork --ip 172.42.0.2 -d deepblue --name blueproxy
clean:
	rm -fv intercept*

all: sign build up


