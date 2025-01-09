#!/bin/bash


echo "/**"
echo " * Beggining the retrieval setup"
echo " */"


echo $TMP_DIR

NumberNodes=$1
NumberUpdates=$2


SLAVES=$(cat other)
MASTER=$(cat bootstrap)

rm -r resultRetrieve
mkdir resultRetrieve
echo "NEW TARRING now :\n===================\n" 
for SLAVE in $SLAVES
do
echo $SLAVE
ssh root@$SLAVE "tar czvf go_trans_$SLAVE.tar.gz CRDT_IPFS/node1/time "
scp root@$SLAVE:~/go_trans_$SLAVE.tar.gz resultRetrieve/go_trans_$SLAVE.tar.gz
scp root@$SLAVE:~/$SLAVE.netlog resultRetrieve/$SLAVE.netlog
scp root@$SLAVE:~/$SLAVE.dstat resultRetrieve/$SLAVE.dstat 
echo "yop"
done 

echo "fordone, do master"
ssh root@$MASTER "sh -c 'tar  czvf go_trans_$MASTER.tar.gz  CRDT_IPFS/node1/time '" 
scp root@$MASTER:~/go_trans_$MASTER.tar.gz resultRetrieve/go_trans_$MASTER.tar.gz 
scp root@$MASTER:~/$MASTER.netlog resultRetrieve/$MASTER.netlog 
scp root@$MASTER:~/$MASTER.dstat resultRetrieve/$MASTER.dstat 

echo "Tarring done, bye bye \n==================\n"  >> taroutside.log


