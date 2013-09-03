#!/bin/sh

set -eu

ar7flashtools () {
	mkdir -p tools

	[ -x tools/srec2bin ] \
		|| gcc -o tools/srec2bin src/openwrt/tools/firmware-utils/src/srec2bin.c
	[ -x tools/addpattern ] \
		|| gcc -o tools/addpattern src/openwrt/tools/firmware-utils/src/addpattern.c
}

patches () {
	mkdir -p src/patches/linux

	ln -s -f $(pwd)/patches/cmdline-parts.patch src/patches/linux/linux-digriz.500-cmdline-parts.patch
	ln -s -f $(pwd)/src/openwrt/target/linux/ar7/patches-3.9/500-serial_kludge.patch src/patches/linux/linux-openwrt.500-serial-kludge.patch
}

buildroot () {
	ln -s -f -T ../../config/buildroot.config src/buildroot/.config

	mkdir -p dl
	ln -s -f -t src/buildroot ../../dl

	export UCLIBC_CONFIG_FILE=../../config/uclibc.config 
	export BUSYBOX_CONFIG_FILE=../../config/busybox.config

	make -C src/buildroot oldconfig
	make -C src/buildroot

	rsync -rl src/buildroot/output/target/ rootfs
}

hostname () {
	echo -n $HOSTNAME > rootfs/etc/hostname

	cat <<EOF > rootfs/etc/hosts
127.0.0.1       localhost
127.1.0.1       $HOSTNAME.$DOMAIN $HOSTNAME

# The following lines are desirable for IPv6 capable hosts
::1     ip6-localhost ip6-loopback
fe00::0 ip6-localnet
ff00::0 ip6-mcastprefix
ff02::1 ip6-allnodes
ff02::2 ip6-allrouters
EOF
}

interfaces () {
	cat <<EOF > rootfs/etc/network/interfaces
# Configure Loopback
auto lo
iface lo inet loopback
	up	modprobe ipv6

	up	sysctl -q -w net.ipv4.conf.default.rp_filter=1
	up	sysctl -q -w net.ipv4.conf.all.rp_filter=1
	up	sysctl -q -w net.ipv4.conf.lo.rp_filter=1

	up	sysctl -q -w net.ipv4.conf.default.forwarding=1
	up	sysctl -q -w net.ipv4.conf.all.forwarding=1
	up	sysctl -q -w net.ipv4.ip_forward=1

	up	iptables-restore  < /etc/network/iptables.active

	up	sysctl -q -w net.netfilter.nf_conntrack_max=4096

	# bogons (rfc6890)
	up	ip route add unreachable 10.0.0.0/8
	up	ip route add unreachable 100.64.0.0/10
	up	ip route add unreachable 169.254.0.0/16
	up	ip route add unreachable 172.16.0.0/12
	up	ip route add unreachable 192.0.0.0/24
	up	ip route add unreachable 192.0.2.0/24
	up	ip route add unreachable 192.88.99.0/24
	up	ip route add unreachable 192.168.0.0/16
	up	ip route add unreachable 198.18.0.0/15
	up	ip route add unreachable 198.51.100.0/24
	up	ip route add unreachable 203.0.113.0/24
	up	ip route add unreachable 240.0.0.0/4
iface lo inet6 static
	address	$WAN6NT::
	netmask	64

	up	sysctl -q -w net.ipv6.conf.default.forwarding=1
	up	sysctl -q -w net.ipv6.conf.all.forwarding=1

	up	sysctl -q -w net.ipv6.conf.default.autoconf=0
	up	sysctl -q -w net.ipv6.conf.all.autoconf=0

	# bogons (rfc6890)
	up	ip route add unreachable 64:ff9b::/96
	up	ip route add unreachable ::ffff:0:0/96
	up	ip route add unreachable 100::/64
	up	ip route add unreachable 2001::/23
	up	ip route add unreachable 2001:2::/48
	up	ip route add unreachable 2001:db8::/32
	up	ip route add unreachable 2001:10::/28
	up	ip route add unreachable 2002::/16
	up	ip route add unreachable fc00::/7

	up	ip6tables-restore < /etc/network/ip6tables.active

auto eth0
iface eth0 inet static
	pre-up	modprobe cpmac
	address	$LAN4IP
	netmask	$LAN4SN
iface eth0 inet6 static
	address	$WAN6NT:$LAN6NT::
	netmask	64
EOF

	if [ "$WAN6IP" != "${WAN6IP#2002:}" ]; then
		cat <<EOF >> rootfs/etc/network/interfaces

# FIXME: we do not get unreachable here
auto tun6to4
iface tun6to4 inet6 v4tunnel
	address $WAN6IP
	netmask 16
	gateway ::192.88.99.1
	endpoint any
	local $WAN4IP
EOF
	fi
}

ppp () {
	:
	# fixups for real
	#up	ip route delete unreachable 2002::/16
	#up	ip route add 2001::/32 dev ppp0
	# fixups for tunnelled
	#up	ip route delete unreachable 192.88.99.0/24
	#up	ip route add 2001::/32 dev tun6to4
}

customise () {
	rsync -rl overlay/ rootfs

	hostname
	interfaces

	find rootfs -type f -name .empty -delete

	rm rootfs/THIS_IS_NOT_YOUR_ROOT_FILESYSTEM

	rm rootfs/etc/init.d/S01logging
	rm rootfs/etc/init.d/S50dropbear

	sed -i -e 's/ext2/auto/' rootfs/etc/fstab
	sed -i -e 's/tmpfs/ramfs/' rootfs/etc/fstab
	sed -i -e 's/^devpts/#devpts/' rootfs/etc/fstab

	# misc unneeded bits
	rm -rf rootfs/home/ftp
	rm -rf rootfs/var/lib
	rm -rf rootfs/var/pcmcia
	rm -rf rootfs/usr/share/udhcpc
	rm -rf rootfs/share/man

	rm -f rootfs/usr/sbin/pppdump
	rm -f rootfs/usr/sbin/pppstats
	rm -f rootfs/usr/sbin/chat

	DELETE="minconn passprompt passwordfd winbind openl2tp pppol2tp"
	for D in $DELETE; do
		rm -f rootfs/usr/lib/pppd/2.4.5/$D.so
	done

	rm -f rootfs/lib/modules/*/source
	rm -f rootfs/lib/modules/*/build
	rm -f rootfs/lib/modules/*/modules.*
	rm -f rootfs/usr/lib/tc/*.dist

	cp -a src/buildroot/output/host/usr/mipsel-buildroot-linux-uclibc/lib/libgcc_s.so* rootfs/lib
	./src/buildroot/output/host/usr/bin/mipsel-linux-sstrip rootfs/lib/libgcc_s.so.1
}

sangam () {
	[ -f rootfs/lib/modules/$BR2_LINUX_KERNEL_VERSION/kernel/drivers/net/tiatm.ko ] && return

	wget -P src -N http://downloads.openwrt.org/sources/sangam_atm-D7.05.01.00-R1.tar.bz2

	rm -rf src/sangam_atm-D7.05.01.00
	tar -xC src -f src/sangam_atm-D7.05.01.00-R1.tar.bz2

	find src/openwrt/package/kernel/ar7-atm/patches-D7.05.01.00 -type f -name '*.patch' \
		| sort \
		| xargs -I{} sh -c "patch -p1 -f -d src/sangam_atm-D7.05.01.00 < '{}'"

	patch -p1 -f -d src/sangam_atm-D7.05.01.00 < patches/sangam_atm.patch

	ARCH=mips make -C src/sangam_atm-D7.05.01.00

	mkdir -p rootfs/lib/firmware
	cp src/sangam_atm-D7.05.01.00/ar0700mp.bin rootfs/lib/firmware/ar0700xx.bin
	cp src/sangam_atm-D7.05.01.00/tiatm.ko rootfs/lib/modules/$BR2_LINUX_KERNEL_VERSION/kernel/drivers/net
}

pppoe () {
	[ -x rootfs/sbin/br2684ctl ] && return

	wget -P src -N --content-disposition http://sourceforge.net/projects/linux-atm/files/latest/download

	tar -xC src -f src/linux-atm-2.5.2.tar.gz

	OLDPWD="$(pwd)"

	cd src/linux-atm-2.5.2

	./configure --prefix="$BASEDIR/rootfs" --with-kernel-headers=$KERNELDIR/include --host=mipsel-linux
	make -C src/lib install
	make -C src/br2684 install

	cd ../../
}

bake () {
	objcopy -S -O srec --srec-forceS3 src/buildroot/output/build/linux-$BR2_LINUX_KERNEL_VERSION/vmlinuz vmlinuz.srec

	tools/srec2bin vmlinuz.srec vmlinuz.bin

	if [ $(wc -c vmlinuz.bin | cut -d' ' -f1) -gt 786432 ]; then
		echo kernel too big
		exit 1
	fi

	DEVTABLE=$(mktemp)

	cat <<'EOF' > "$DEVTABLE"
# <name>	<type>	<mode>	<uid>	<gid>	<major>	<minor>	<start>	<inc>	<count>
/bin/busybox	f	4755	0	0	-	-	-	-	-
EOF

	/usr/sbin/mkfs.jffs2 -D "$DEVTABLE" -X zlib -x lzo -x rtime -e 65536 -n -p -t -l -d rootfs --squash -o fs.img
	if [ $(wc -c fs.img | cut -d' ' -f1) -gt 3211264 ]; then
		echo filesystem too big
		exit 1
	fi

	rm "$DEVTABLE"

	( dd if=/dev/zero bs=16 count=1; dd if=vmlinuz.bin bs=786432 conv=sync; cat fs.img ) | tools/addpattern -o firmware-code.bin -p WA21
}

if [ -f local ]; then
	. ./local
else
	echo missing 'local' config file, default created, please edit >&2
	cat <<'EOF' > local
HOSTNAME=host
DOMAIN=example.com

LAN4IP=192.168.1.1
LAN4SN=255.255.255.0

WAN4IP=203.0.113.1
# put here your /48 IPv6 allocation if you have one,
# like "WAN6NT=2001:db8:beef", otherwise leave blank
WAN6NT=

if [ ! "$WAN6NT" ]; then
	WAN6NT=$(echo $WAN4IP | tr . ' ' | xargs printf '2002:%.2x%.2x:%.2x%.2x')
fi

WAN6IP=$WAN6NT::
# LAN IPv6 subnet to use (WAN6NT:LAN6NT::/64)
LAN6NT=1000

# 'PPPoE' or 'PPPoA'
TYPE=PPPoA
USER=username
PASS=password
EOF
	exit 1
fi

git submodule init
git submodule update

BASEDIR="$(pwd)"

ar7flashtools

patches

buildroot

eval $(grep BR2_LINUX_KERNEL_VERSION src/buildroot/.config)
export KERNELDIR="$BASEDIR/src/buildroot/output/build/linux-$BR2_LINUX_KERNEL_VERSION"
export CROSS_COMPILE="$BASEDIR/src/buildroot/output/host/usr/bin/mipsel-linux-"

[ "$TYPE" = "PPPoE" ] && pppoe
sangam

customise

bake

echo
echo "your firmware is now ready to deploy, do this by typing:"
echo "echo -e \"mode binary\\\nconnect $LAN4IP\\\nput firmware-code.bin\" | tftp"

exit 0
