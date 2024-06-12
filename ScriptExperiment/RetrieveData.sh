#!/bin/bash

inputfolder=$1
savingfolder=$2


for filename in $inputfolder/*.tar.gz; do

    EXTENSION=`echo "$filename" | cut -d'.' -f1`
    EXTENSION=`echo "$EXTENSION" | cut -d'/' -f2`
    echo "$savingfolder/$EXTENSION"
    echo "filename: $filename"
    mkdir "$savingfolder/$EXTENSION"
    tar -xf $filename -C $savingfolder/$EXTENSION > /dev/null
done

mv $inputfolder/*.netlog  $savingfolder
mv $inputfolder/*.dstat  $savingfolder
