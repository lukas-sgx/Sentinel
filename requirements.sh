#!/bin/sh

apk add --no-cache bash curl tar shadow sudo bash wget git
wget https://go.dev/dl/go1.25.3.linux-amd64.tar.gz
tar -C /usr/local -xzf go1.25.3.linux-amd64.tar.gz
rm go1.25.3.linux-amd64.tar.gz
export PATH=$PATH:/usr/local/go/bin