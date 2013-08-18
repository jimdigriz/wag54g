#!/bin/sh

set -eu

PPPOE=
USAGE="Usage: $(basename $0) [OPTION]

  -e                   enable PPP-over-Ethernet support
  -h                   display this help and exit
"

while getopts eh f
do
	case $f in
	e)	PPPOE=$f;;
	h | \?)	echo "$USAGE"; [ $f = 'h' ] && exit 0 || exit 1;;
	esac
done
shift $(expr $OPTIND - 1)

ar7flashtools () {
	[ -x tools/srec2bin ] \
		|| gcc -o tools/srec2bin src/openwrt/tools/firmware-utils/src/srec2bin.c
	[ -x tools/addpattern ] \
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

	ln -s -f -T ../../src/openwrt/target/linux/ar7/patches-3.9/500-serial_kludge.patch patches/linux/linux-openwrt.500-serial-kludge.patch
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

	make -C src/sangam_atm-D7.05.01.00
}

pppoe () {
	wget -P src -N --content-disposition http://sourceforge.net/projects/linux-atm/files/latest/download

	tar -xC src -f src/linux-atm-2.5.2.tar.gz

	cd src/linux-atm-2.5.2

	./configure --with-kernel-headers=$KERNELDIR/include --host=mipsel-linux
	make -C src/lib
	make -C src/br2684

	cd ../..

	#$ cp ../../../linux-atm/src/br2684/.libs/br2684ctl usr/sbin
	#$ cp ../../../linux-atm/src/lib/.libs/libatm.so.1.0.0 usr/lib
	#$ /usr/src/wag54g/buildroot/output/host/usr/bin/mipsel-linux-sstrip usr/sbin/br2684ctl
	#$ /usr/src/wag54g/buildroot/output/host/usr/bin/mipsel-linux-sstrip usr/lib/libatm.so.1.0.0
	#$ ln -s libatm.so.1.0.0 usr/lib/libatm.so
	#$ ln -s libatm.so.1.0.0 usr/lib/libatm.so.1
}

git submodule init
git submodule update

#VERSION_OPENWRT=$(git --git-dir=src/openwrt/.git rev-parse HEAD | cut -c 1-8)

mkdir -p tools

ar7flashtools

#patches

buildroot

export KERNELDIR="$(pwd)/src/buildroot/output/build/linux-3.10.7"
export CROSS_COMPILE="$(pwd)/src/buildroot/output/host/usr/bin/mipsel-linux-"

[ "$PPPOE" ] && pppoe
#sangam

exit 0
