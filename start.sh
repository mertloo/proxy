#!/bin/bash
if [[ $1 == "socks5" ]];then
	./proxy -next 0.0.0.0:20002 -password woshimima -timeout 2 -goro 5
elif [[ $1 == "ssocks" ]];then
	./proxy -proto ssocks -addr 0.0.0.0:20002 -password woshimima -timeout 2 -goro 4
else
	echo 'Usage: start.sh <socks5|ssocks>'
fi
