/sbin/iptables -t nat -A POSTROUTING -o "$2" -j MASQUERADE
/sbin/iptables -A FORWARD -i "$2" -o "$1" -m state --state RELATED,ESTABLISHED -j ACCEPT
/sbin/iptables -A FORWARD -i "$1" -o "$2" -j ACCEPT