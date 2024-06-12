#!/bin/bash
code=$1
echo "/**"
echo " * Beggining the CRDT IPFS setup"
echo " */"

USER_LOGIN_NAME=$(id -un)
USER_GROUP_ID=$(id -g)
USER_GROUP_NAME=$(id -gn)

TMP_DIR=/tmp/$DATE'-'$$'-CRDTIPFS'

echo $TMP_DIR

SLAVES=$(cat other)
MASTER=$(cat bootstrap)


echo  "SLAVES"
echo $SLAVES
echo "MASTER"
echo $MASTER
echo "Building the  GO implementation"

for SLAVE in $SLAVES
do
scp $code root@$SLAVE:~/$code &
done
scp $code root@$MASTER:~/$code
sleep 20s


for SLAVE in $SLAVES
do
ssh root@$SLAVE "sh -c 'tar -xvf go_trans.tar.gz > tar.log && cd CRDT_IPFS && /usr/local/go/bin/go mod tidy > build.log && /usr/local/go/bin/go build > build.log'" > /dev/null  2>&1 & 
done
ssh root@$MASTER "sh -c 'tar -xvf go_trans.tar.gz > tar.log && cd CRDT_IPFS && /usr/local/go/bin/go mod tidy > build.log && /usr/local/go/bin/go build > build.log'" > build.log 2>&1

sleep 100s