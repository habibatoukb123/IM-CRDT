ldisk=/dev/$(lsblk | grep -w 186.3G |  head -n1  | awk '{print $1;}')
mkfs $ldisk
mount $ldisk /home