#!/bin/sh
wget https://go.dev/dl/go1.19.4.linux-amd64.tar.gz -O go1.19.4.linux-amd64.tar.gz

rm -rf /usr/local/go && tar -C /usr/local -xzf go1.19.4.linux-amd64.tar.gz
export PATH=$PATH:/usr/local/go/bin

rm go1.19.4.linux-amd64.tar.gz
