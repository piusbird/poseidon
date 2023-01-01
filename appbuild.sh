#!/bin/sh

git clone https://git.piusbird.space/miniweb.git/
cd miniweb
make
cp miniwebproxy /app
cp -r scripts/ /app
cd ..
sh -c miniweb/sign.sh
git clone https://git.piusbird.space/poseidon.git/
cd poseidon 
go build
cp poseidon /app
