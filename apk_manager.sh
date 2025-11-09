#!/bin/sh

set -e

ROOTFS="${1:-rootfs}"

if [ ! -d "$ROOTFS" ]; then
    echo "❌ Le dossier rootfs '$ROOTFS' n'existe pas."
    echo "Usage : ./restore_apk.sh /chemin/vers/rootfs"
    exit 1
fi

echo "➡️ Téléchargement du minirootfs Alpine..."
wget https://dl-cdn.alpinelinux.org/alpine/latest-stable/releases/x86_64/alpine-minirootfs-3.22.0-x86_64.tar.gz -O alpine-minirootfs-latest-x86_64.tar.gz

echo "➡️ Extraction temporaire..."
mkdir -p alpine_tmp
tar -xzf alpine-minirootfs-latest-x86_64.tar.gz -C alpine_tmp

echo "➡️ Copie de apk..."
mkdir -p "$ROOTFS/sbin"
cp alpine_tmp/sbin/apk "$ROOTFS/sbin/"

echo "➡️ Copie des clés APK..."
mkdir -p "$ROOTFS/etc"
cp -r alpine_tmp/etc/apk "$ROOTFS/etc/"

echo "➡️ Copie des librairies nécessaires..."
mkdir -p "$ROOTFS/usr/lib"
cp alpine_tmp/usr/lib/libapk* "$ROOTFS/usr/lib/" 2>/dev/null || true
cp alpine_tmp/usr/lib/libcrypto* "$ROOTFS/usr/lib/" 2>/dev/null || true
cp alpine_tmp/usr/lib/libssl* "$ROOTFS/usr/lib/" 2>/dev/null || true
cp alpine_tmp/usr/lib/libz* "$ROOTFS/usr/lib/" 2>/dev/null || true

echo "➡️ Nettoyage..."
rm -rf alpine_tmp
rm alpine-minirootfs-latest-x86_64.tar.gz

echo "✅ apk restored !"

echo "➡️ Test chroot..."
mount --bind /dev "$ROOTFS/dev" 2>/dev/null || true

if chroot "$ROOTFS" /sbin/apk --version; then
    echo "✅ apk fonctionne dans le rootfs."
else
    echo "❌ apk ne fonctionne pas (manque peut-être une lib)."
fi
