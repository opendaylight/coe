until ifconfig br-int > /dev/null
do
  sleep 1
done

# give odl time to configure the bridge
sleep 10

BRIDGE_IP=$(ip a show br-int | awk '/inet/{print substr($2,0)}' | sed 's/\/.*//' | sed 's/1$/254/')
echo $BRIDGE_IP
