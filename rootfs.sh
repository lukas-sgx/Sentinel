#!/bin/sh

wget https://dl-cdn.alpinelinux.org/alpine/latest-stable/releases/x86_64/alpine-minirootfs-3.22.0-x86_64.tar.gz
mkdir rootfs
tar -xzf alpine-minirootfs-3.22.0-x86_64.tar.gz -C rootfs
rm alpine-minirootfs-3.22.0-x86_64.tar.gz
cd rootfs
echo "nameserver 1.1.1.1" > etc/resolv.conf
# find . -type l -exec sh -c '
#   for l; do
#     target=$(readlink "$l")
#     case "$target" in
#       /bin/busybox|/usr/bin/busybox|/sbin/busybox|/usr/sbin/busybox)
#         ln -sf ./busybox "$l"
#     esac
#   done
# ' sh {} +