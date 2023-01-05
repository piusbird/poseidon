build: sign
	docker build --no-cache . -t deepblue
sign:
	./sign.sh
up: sign 
	docker run --net customnetwork --ip 172.42.0.2 -d deepblue --name blueproxy
clean:
	rm -fv intercept*
	rm -rfv app
nodocker:
	mkdir -pv app/src
	cp appbuild.sh app/src
	cp intercept.* app
	cd app/src
	export appRoot=` realpath ../app`
	./appbuild.sh

all: sign build up


