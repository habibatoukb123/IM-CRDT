NumberNodes=$1
NumberUpdates=$2
nbpeersUpdating=$3
file_NODES=$4
waitTime=$5
SyncTime=$6


echo "start Kadeploy "$(date +"%T")
kadeploy3 -a CRDT_IPFS.yaml > kadeploy.log ;
echo "kadeploy3 done "$(date +"%T")

echo "NumberNodes : $NumberNodes"
cat "$file_NODES" | cut -d'.' -f1 | tail -n "$(( $NumberNodes - 1 ))" > other
head -n 1 "$file_NODES"| cut -d'.' -f1 > bootstrap

SLAVES=$(cat other)
MASTER=$(cat bootstrap)

./CompileNodes.sh go_trans.tar.gz
./Run_CI.sh $NumberNodes $NumberUpdates $nbpeersUpdating $file_NODES $waitTime $SyncTime

echo "Every peers has been started, starting iftop"

sleeptime=$(( $NumberUpdates ))
margintime=$(( 200 ))

for SLAVE in $SLAVES
do
    scp get_network_csv.sh root@$SLAVE:~/get_network_csv.sh
    ssh root@$SLAVE "sh -c './get_network_csv.sh $(($sleeptime + $margintime *3 / 4 )) ${SLAVE}.netlog' " 2>&1 > /dev/null &
    #ssh root@$SLAVE "sh -c 'iftop -t -s $(($sleeptime + $margintime / 2 ))  > ${SLAVE}.netlog'" 2>&1 > /dev/null &
    ssh root@$SLAVE "sh -c 'dstat -tcnmdsp 3 > ${SLAVE}.dstat'" 2>&1 > /dev/null &
done
scp get_network_csv.sh root@$MASTER:~/get_network_csv.sh
ssh root@$MASTER "sh -c './get_network_csv.sh $(($sleeptime + $margintime *3 / 4 )) ${MASTER}.netlog' " 2>&1 > /dev/null &
#ssh root@$MASTER "sh -c 'iftop -t -s $(($sleeptime + $margintime / 2 ))  > ${MASTER}.netlog'" 2>&1 > /dev/null &
ssh root@$MASTER "sh -c 'dstat -tcnmdp 3 > ${MASTER}.dstat'" 2>&1 > /dev/null &

echo "waiting 60s so everybody is connected"


stepsleeptime=$(( ($sleeptime + $margintime)/10 ))
sleep 60s
echo "sleepTime : $sleeptime = $NumberUpdates * $SyncTime"
echo "letting the algorithm run for $(( $stepsleeptime * 10 ))"
echo "i.e. sleeping 10 times this nb of seconds :  $(( $stepsleeptime ))"
for percent in {1..100..10} 
do
    echo $(( $percent ))"percent"
    sleep $(( $stepsleeptime ))s
done
#./test_file.sh

./retrieveInfo.sh $NumberNodes $NumberUpdates
sleep 2s
echo "DONE, now retrieve  data and mean!!!"

