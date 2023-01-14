#!/bin/sh
if [ ! -n $appRoot ]
then
	export appRoot="/app"
fi
mkdir -pv $appRoot/src
cd $appRoot/src
echo $appRoot
git clone https://git.piusbird.space/miniweb.git/
cd miniweb
make
cp miniwebproxy $appRoot
cp -r scripts/ $appRoot
chmod +x sign.sh
./sign.sh
cp intercept* $appRoot
cd ..
sh -c miniweb/sign.sh
git clone https://git.piusbird.space/poseidon.git/
cd poseidon 
go build
cp *.html $appRoot
cp -r assets $appRoot
cp poseidon $appRoot
