#!/bin/sh

set -eu

ar7flashtools () {
	[ -x tools/srec2bin ] \
		|| gcc -o tools/srec2bin src/openwrt/tools/firmware-utils/src/srec2bin.c
	[ -x tools/addpatten ] \
		|| gcc -o tools/addpattern src/openwrt/tools/firmware-utils/src/addpattern.c
}

buildroot () {
	ln -s -f -T ../../config/buildroot.config src/buildroot/.config

	mkdir -p dl
	ln -s -f -t src/buildroot ../../dl

	export UCLIBC_CONFIG_FILE=../../config/uclibc.config 
	export BUSYBOX_CONFIG_FILE=../../config/busybox.config

	make -C src/buildroot oldconfig
	make -C src/buildroot
}

git submodule init
git submodule update

mkdir -p tools

ar7flashtools

buildroot

exit 0
