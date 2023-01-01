#!/bin/sh
cd /app
./miniwebproxy >> /stackmsg 2>&1 &
./poseidon >> /stackmsg 2>&1
tail -f /stackmsg

