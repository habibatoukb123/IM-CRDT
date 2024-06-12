

SLAVES=$(cat other)
MASTER=$(cat bootstrap)


for SLAVE in $SLAVES
do
scp Testg5K.tar.gz root@$SLAVE:~/Testg5K.tar.gz &
done
scp Testg5K.tar.gz root@$MASTER:~/Testg5K.tar.gz
sleep 20s


for SLAVE in $SLAVES
do
ssh root@$SLAVE "sh -c 'tar xf Testg5K.tar.gz && cd Testg5K && /usr/local/go/bin/go build && ./Testg5k '" &
done
ssh root@$MASTER "sh -c 'tar xf Testg5K.tar.gz && cd Testg5K && /usr/local/go/bin/go build && ./Testg5k '"
sleep 20s



