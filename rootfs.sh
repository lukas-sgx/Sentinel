#!/bin/sh

wget https://dl-cdn.alpinelinux.org/alpine/latest-stable/releases/x86_64/alpine-minirootfs-3.22.0-x86_64.tar.gz
mkdir rootfs
tar -xzf alpine-minirootfs-3.22.0-x86_64.tar.gz -C rootfs
rm alpine-minirootfs-3.22.0-x86_64.tar.gz
cd rootfs
echo "nameserver 1.1.1.1" > etc/resolv.conf