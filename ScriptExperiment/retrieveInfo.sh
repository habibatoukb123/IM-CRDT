#!/bin/bash


echo "/**"
echo " * Beggining the retrieval setup"
echo " */"


echo $TMP_DIR

NumberNodes=$1
NumberUpdates=$2


let NHOST=($(cat $OAR_FILE_NODES | uniq | wc -l)-1)
SLAVES=$(cat other)
MASTER=$(cat bootstrap)

rm -r resultRetrieve
mkdir resultRetrieve
echo "NEW TARRING now :\n===================\n"  >> taroutside.log
for SLAVE in $SLAVES
do
ssh root@$SLAVE "sh -c 'tar czvf go_trans_$SLAVE.tar.gz CRDT_IPFS/node1/time' " >> taroutside.log
scp  root@$SLAVE:~/go_trans_$SLAVE.tar.gz resultRetrieve/go_trans_$SLAVE.tar.gz   >> taroutside.log
scp root@$SLAVE:~/$SLAVE.netlog resultRetrieve/$SLAVE.netlog >> taroutside.log
scp root@$SLAVE:~/$SLAVE.dstat resultRetrieve/$SLAVE.dstat >> taroutside.log
done 

ssh root@$MASTER "sh -c 'tar  czvf go_trans_$MASTER.tar.gz  CRDT_IPFS/node1/time '"  >> taroutside.log
scp root@$MASTER:~/go_trans_$MASTER.tar.gz resultRetrieve/go_trans_$MASTER.tar.gz  >> taroutside.log
scp root@$MASTER:~/$MASTER.netlog resultRetrieve/$MASTER.netlog  >> taroutside.log
scp root@$MASTER:~/$MASTER.dstat resultRetrieve/$MASTER.dstat  >> taroutside.log

echo "Tarring done, bye bye \n==================\n"  >> taroutside.log


