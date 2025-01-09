#!/bin/bash
numberPeers=$1
numberUpdates=$2
nbpeersUpdating=$3
file_NODES=$4
waitTime=$5
SyncTime=$6

echo "/**"
echo " * Beggining the CRDT IPFS setup"
echo " */"

USER_LOGIN_NAME=$(id -un)
USER_GROUP_ID=$(id -g)
USER_GROUP_NAME=$(id -gn)

numberPeers=$2
DATE=$(date +%s)

TMP_DIR=/tmp/$DATE'-'$$'-CRDTIPFS'

echo $TMP_DIR

SLAVES=$(cat other)
MASTER=$(cat bootstrap)


echo  "SLAVES"
echo $SLAVES
echo "MASTER"
echo $MASTER
echo "Building the  GO implementation"


echo "running the bootstrap in ${MASTER} node1"
ssh root@$MASTER "rm  CRDT_IPFS/ID"



ssh root@$MASTER "mkdir  CRDT_IPFS/node1"

ssh root@$MASTER "sh -c 'sysctl -w net.ipv6.conf.all.disable_ipv6=1 && sysctl -w net.ipv6.conf.default.disable_ipv6=1 '"
# ssh root@$MASTER "sh -c 'cd CRDT_IzPFS && ./IPFS_CRDT --encode sataislifesataisloveanditsfor32b --mode BootStrap --name node1 --updatesNB $numberUpdates --updating true  > /dev/null & '"  &


#for test ./IPFS_CRDT --mode=BootStrap --name=node1 --updatesNB=300 --updating=true --WaitTime=30 --SyncTime=1
#ssh root@$MASTER "sh -c 'export LIBP2P_FORCE_PNET=1 && cd CRDT_IPFS && ./IPFS_CRDT  --mode BootStrap --name node1 --updatesNB $numberUpdates --updating true  > /dev/null & '"  &
# ssh root@$MASTER "sh -c 'cd CRDT_IPFS && ./IPFS_CRDT --mode=BootStrap --ParallelRetrieve=0 --name=node1 --updatesNB=$numberUpdates --updating=true --WaitTime=$waitTime --SyncTime=$SyncTime  > /dev/null & '"  &
ssh root@$MASTER "sh -c 'cd CRDT_IPFS && ./IPFS_CRDT --ParallelRetrieve=1  --mode=BootStrap --name=node1 --updatesNB=$numberUpdates --updating=true --WaitTime=$waitTime   --SyncTime=$SyncTime  > /dev/null & '"  &

sleep 30s
BOOTSTRAPIDS=$(ssh root@$MASTER "sh -c 'cat ./CRDT_IPFS/ID2'")
BOOTSTRAPID=""
#echo "reading file $BOOTSTRAPIDS" ./IPFS_CRDT  --mode BootStrap --name node1 --updatesNB 100 --updating true
for ID in $BOOTSTRAPIDS
do
#echo "analysing $ID"
if [[ "$ID" == *"/ip4"* ]]; then
  if [[ "$ID" == *"/127.0.0"* ]]; then
#    "It's a Local IP, not interesting"
    continue
  else
#    "bootstrap IP is usable"
    BOOTSTRAPID=$ID
  fi
fi

done
echo "running the lisnteners -- FIRST"
x=$(( $nbpeersUpdating - 1 ))
echo "x: "$x


### Sending the IP of the bootstrap to all other for IPFS and libp2p pubsub ###
scp root@$MASTER:~/CRDT_IPFS/IDBootstrapIPFS IDBootstrapIPFS
sleep 5s 
for SLAVE in $SLAVES
do
scp IDBootstrapIPFS root@$SLAVE:~/CRDT_IPFS/IDBootstrapIPFS
done

sleep 10s


for SLAVE in $SLAVES
do


ssh root@$SLAVE "rm -rf CRDT_IPFS/node1"
ssh root@$SLAVE "mkdir  CRDT_IPFS/node1"


echo $SLAVE

#ssh root@$SLAVE "rm -rf CRDT_IPFS/node2"
#ssh root@$SLAVE "rm -rf CRDT_IPFS/node3"
#ssh root@$SLAVE "rm -rf CRDT_IPFS/node4"



ssh root@$SLAVE "sh -c 'sysctl -w net.ipv6.conf.all.disable_ipv6=1 && sysctl -w net.ipv6.conf.default.disable_ipv6=1 '"
if [[ $x > 0 ]]
then
echo "updating"

# ssh root@$SLAVE "sh -c 'cd CRDT_IPFS && ./IPFS_CRDT --encode sataislifesataisloveanditsfor32b --mode update --ni ${BOOTSTRAPID} --name node1 --updatesNB $numberUpdates --updating true  > /dev/null &'" &
# 
# ./IPFS_CRDT --mode update --ni /ip4/172.16.193.5/udp/42911/quic-v1/webtransport/certhash/uEiBRLcQZ0wJ5qbdbXiOWnZ7e-NNCm6bAxHnQwIYspQrVag/certhash/uEiCtpQctH5sNZjsSLmU1u7_gE4DlYCJDf4dwccm01RhVsQ/p2p/12D3KooWC5y4WAcM2yxb1LtV2F3b253zEuqtuuV7tUUPnKi1dLyN --name node1 --updatesNB 100 --IPFSBootstrap ~/CRDT_IPFS/IDBootstrapIPFS --updating true 
ssh root@$SLAVE "sh -c 'cd CRDT_IPFS  && ./IPFS_CRDT  --mode=update --ParallelRetrieve=1 --ni=${BOOTSTRAPID} --name=node1 --updatesNB=$numberUpdates  --IPFSBootstrap=IDBootstrapIPFS --updating=true --WaitTime=$waitTime  --SyncTime=$SyncTime  > /dev/null &'" &
#ssh root@$SLAVE "sh -c 'cd CRDT_IPFS  && ./IPFS_CRDT  --ParallelRetrieve=1 --mode=update --ni=${BOOTSTRAPID} --name=node1 --updatesNB=$numberUpdates  --IPFSBootstrap=IDBootstrapIPFS --updating=true --WaitTime=$waitTime  --SyncTime=$SyncTime  > /dev/null &'" &
x=$(( $x - 1 ))
else
echo "NOT updating"
# ssh root@$SLAVE "sh -c 'cd CRDT_IPFS && ./IPFS_CRDT --encode sataislifesataisloveanditsfor32b --mode update --ni ${BOOTSTRAPID} --name node1 --updatesNB $numberUpdates  > /dev/null &'" &
# ssh root@$SLAVE "sh -c 'cd CRDT_IPFS && ./IPFS_CRDT --mode=update --ParallelRetrieve=0 --ni=${BOOTSTRAPID} --name=node1 --updatesNB=$numberUpdates  --IPFSBootstrap=IDBootstrapIPFS --WaitTime=$waitTime  --SyncTime=$SyncTime  > /dev/null &'" &
#ssh root@$SLAVE "sh -c 'cd CRDT_IPFS && ./IPFS_CRDT --mode=update --ParallelRetrieve=0 --ni="/ip4/172.16.96.68/tcp/35293/p2p/12D3KooWE1Gd4ZvTSbqg9iSWbcVAnAej3zRqcEW9SwDv7jEZoduW" --name=node1 --updatesNB=10  --IPFSBootstrap=IDBootstrapIPFS --WaitTime=500  --SyncTime=1  > /dev/null &'" &
ssh root@$SLAVE "sh -c 'cd CRDT_IPFS && ./IPFS_CRDT --ParallelRetrieve=1 --mode=update --ni=${BOOTSTRAPID} --name=node1 --updatesNB=$numberUpdates  --IPFSBootstrap=IDBootstrapIPFS --WaitTime=$waitTime  --SyncTime=$SyncTime  > /dev/null &'" &
fi

#ssh root@$SLAVE "sh -c 'cd CRDT_IPFS && ./IPFS_CRDT --mode update --ni ${BOOTSTRAPID} --name node2 > out.log &'"&
#ssh root@$SLAVE "sh -c 'cd CRDT_IPFS && ./IPFS_CRDT --mode update --ni ${BOOTSTRAPID} --name node3 > out.log &'"&
#ssh root@$SLAVE "sh -c 'cd CRDT_IPFS && ./IPFS_CRDT --mode update --ni ${BOOTSTRAPID} --name node4 > out.log &'"&

done
echo "All listeners have been started"





#saves Working :
#root@ecotype-48:~/CRDT_IPFS# cat IDBootstrapIPFSSave_Working 
#{"ID":"QmRpTkLCxi6aUjQ1WKZJZY8o3eCoxVFjMeffAkquT1SfrZ","Addrs":["/ip6/::1/udp/4001/quic","/ip6/::1/udp/4001/quic-v1/webtransport/certhash/uEiBis2ou30toWGelAt12UgDmEnPn6Ih06OTdFbsp1PRxxg/certhash/uEiALLUvk1LdXWyjxBCe1VLxQOc6jgvgVm2_KKJXL7qEkNw","/ip6/::1/tcp/4001","/ip4/127.0.0.1/udp/4001/quic","/ip4/127.0.0.1/udp/4001/quic-v1","/ip4/172.16.193.46/udp/4001/quic-v1","/ip4/172.16.193.46/udp/4001/quic-v1/webtransport/certhash/uEiBis2ou30toWGelAt12UgDmEnPn6Ih06OTdFbsp1PRxxg/certhash/uEiALLUvk1LdXWyjxBCe1VLxQOc6jgvgVm2_KKJXL7qEkNw","/ip4/127.0.0.1/udp/4001/quic-v1/webtransport/certhash/uEiBis2ou30toWGelAt12UgDmEnPn6Ih06OTdFbsp1PRxxg/certhash/uEiALLUvk1LdXWyjxBCe1VLxQOc6jgvgVm2_KKJXL7qEkNw","/ip6/::1/udp/4001/quic-v1","/ip4/172.16.193.46/tcp/4001","/ip4/127.0.0.1/tcp/4001","/ip4/172.16.193.46/udp/4001/quic"]}

#{"ID":"QmWSftxqfvmcaAnnrprQhPrwErvz9hWpWgo8DRPxGCx4VU","Addrs":["/ip6/::1/udp/4001/quic-v1","/ip4/127.0.0.1/tcp/4001","/ip4/127.0.0.1/udp/4001/quic-v1/webtransport/certhash/uEiChTy1J6oRyOe32JVLgddZZ2wpFuDt7x5J5s-3g81SSKg/certhash/uEiB10JPSKJKBTzMo3jpfWOtjQOWWKqUy7FfQ2oV857gjGg","/ip4/172.16.193.46/udp/4001/quic","/ip4/127.0.0.1/udp/4001/quic","/ip4/127.0.0.1/udp/4001/quic-v1","/ip6/::1/udp/4001/quic-v1/webtransport/certhash/uEiChTy1J6oRyOe32JVLgddZZ2wpFuDt7x5J5s-3g81SSKg/certhash/uEiB10JPSKJKBTzMo3jpfWOtjQOWWKqUy7FfQ2oV857gjGg","/ip4/172.16.193.46/tcp/4001","/ip4/172.16.193.46/udp/4001/quic-v1/webtransport/certhash/uEiChTy1J6oRyOe32JVLgddZZ2wpFuDt7x5J5s-3g81SSKg/certhash/uEiB10JPSKJKBTzMo3jpfWOtjQOWWKqUy7FfQ2oV857gjGg","/ip6/::1/udp/4001/quic","/ip6/::1/tcp/4001","/ip4/172.16.193.46/udp/4001/quic-v1"]}



#not working :
# {"ID":"Qmckrd6FjUmS8Rq9tdyRtnsFpKjw8ShkJzj3JPwtqYENYG","Addrs":["/ip4/172.16.193.48/udp/4001/quic-v1","/ip4/127.0.0.1/udp/4001/quic-v1","/ip6/::1/udp/4001/quic","/ip4/172.16.193.48/tcp/4001","/ip6/::1/tcp/4001","/ip4/172.16.193.48/udp/4001/quic","/ip6/::1/udp/4001/quic-v1/webtransport/certhash/uEiBuOtE7xVeCcqx0C_-caSLe-KO8OfI8O6oSVz5TmKfmzw/certhash/uEiCtj_gCX7NK5_YTqtjCGby1ja3-h6xEgN18J4GOe2FVhg","/ip4/127.0.0.1/udp/4001/quic","/ip6/::1/udp/4001/quic-v1","/ip4/127.0.0.1/tcp/4001","/ip4/172.16.193.48/udp/4001/quic-v1/webtransport/certhash/uEiBuOtE7xVeCcqx0C_-caSLe-KO8OfI8O6oSVz5TmKfmzw/certhash/uEiCtj_gCX7NK5_YTqtjCGby1ja3-h6xEgN18J4GOe2FVhg","/ip4/127.0.0.1/udp/4001/quic-v1/webtransport/certhash/uEiBuOtE7xVeCcqx0C_-caSLe-KO8OfI8O6oSVz5TmKfmzw/certhash/uEiCtj_gCX7NK5_YTqtjCGby1ja3-h6xEgN18J4GOe2FVhg"]}

#{"ID":"QmNrZyFc9r3R6oL3Kcz8i4QzPb1dUK7dCQQLPi5BUb5LrT","Addrs":["/ip6/::1/tcp/4001","/ip4/172.16.193.48/udp/4001/quic","/ip4/127.0.0.1/udp/4001/quic","/ip4/172.16.193.48/udp/4001/quic-v1","/ip4/127.0.0.1/udp/4001/quic-v1","/ip6/::1/udp/4001/quic","/ip6/::1/udp/4001/quic-v1/webtransport/certhash/uEiAId3b0mluAedDSk2u_gqSyQw9WB36Rr1L6wWG0wTe-CQ/certhash/uEiBd98ziuq3OY4jenlN7tC6Vrva2NHKd12bYNsOdjaKUkQ","/ip4/127.0.0.1/tcp/4001","/ip4/127.0.0.1/udp/4001/quic-v1/webtransport/certhash/uEiAId3b0mluAedDSk2u_gqSyQw9WB36Rr1L6wWG0wTe-CQ/certhash/uEiBd98ziuq3OY4jenlN7tC6Vrva2NHKd12bYNsOdjaKUkQ","/ip4/172.16.193.48/udp/4001/quic-v1/webtransport/certhash/uEiAId3b0mluAedDSk2u_gqSyQw9WB36Rr1L6wWG0wTe-CQ/certhash/uEiBd98ziuq3OY4jenlN7tC6Vrva2NHKd12bYNsOdjaKUkQ","/ip6/::1/udp/4001/quic-v1","/ip4/172.16.193.48/tcp/4001"]}
