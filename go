#!/bin/sh

set -eu

ar7flashtools () {
	[ -x tools/srec2bin ] \
		|| gcc -o tools/srec2bin src/openwrt/tools/firmware-utils/src/srec2bin.c
	[ -x tools/addpatten ] \
		|| gcc -o tools/addpattern src/openwrt/tools/firmware-utils/src/addpattern.c
}

patches () {
	mkdir -p patches/linux

	#alex@berk:/usr/src/wag54g/wag54g$ find src/openwrt/target/linux/ar7 -name '*.patch'
	#src/openwrt/target/linux/ar7/patches-3.9/110-flash.patch
	#src/openwrt/target/linux/ar7/patches-3.9/120-gpio_chrdev.patch
	#src/openwrt/target/linux/ar7/patches-3.9/160-vlynq_try_remote_first.patch
	#src/openwrt/target/linux/ar7/patches-3.9/200-free-mem-below-kernel-offset.patch
	#src/openwrt/target/linux/ar7/patches-3.9/300-add-ac49x-platform.patch
	#src/openwrt/target/linux/ar7/patches-3.9/310-ac49x-prom-support.patch
	#src/openwrt/target/linux/ar7/patches-3.9/320-ac49x-mtd-partitions.patch
	#src/openwrt/target/linux/ar7/patches-3.9/500-serial_kludge.patch
	#src/openwrt/target/linux/ar7/patches-3.9/920-ar7part.patch
	#src/openwrt/target/linux/ar7/patches-3.9/925-actiontec_leds.patch
	#src/openwrt/target/linux/ar7/patches-3.9/950-cpmac_titan.patch
	#src/openwrt/target/linux/ar7/patches-3.9/972-cpmac_fixup.patch

	ln -s -f -T ../../src/openwrt/target/linux/ar7/patches-3.9/500-serial_kludge.patch patches/linux/openwrt.500-serial-kludge.patch
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

sangam () {
	wget -P src -N http://downloads.openwrt.org/sources/sangam_atm-D7.05.01.00-R1.tar.bz2

	rm -rf src/sangam_atm-D7.05.01.00
	tar -xC src -f src/sangam_atm-D7.05.01.00-R1.tar.bz2

	find src/openwrt/package/kernel/ar7-atm/patches-D7.05.01.00 -type f -name '*.patch' \
		| xargs -I{} sh -c "patch -p1 -f -d src/sangam_atm-D7.05.01.00 < '{}'"
}

git submodule init
git submodule update

VERSION_OPENWRT=$(git --git-dir=src/openwrt/.git rev-parse HEAD | cut -c 1-8)

mkdir -p tools

ar7flashtools

#patches

buildroot

sangam

exit 0
