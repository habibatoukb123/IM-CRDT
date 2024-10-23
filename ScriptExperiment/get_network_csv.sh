#!/bin/bash

experimentation_time=$1
filename=$2

# ifstat can sum the bandwidth across specified interfaces, which it displays
# as an additional column of output. This script trims the output to contain
# only the Total column.

# For debugging, uncomment the following line
# set -x

# First execute ifstat once to find the number of interfaces we'll be watching.
# We pass through any command line arguments to allow the user to select a
# subset of interfaces if desired. Capture the first line of output, which is
# the header row that lists interface names. Count the words and trim the
# whitespace out of the 'wc' output.
IFCOUNT=$(ifstat | head -n 1 | wc -w | tr -d ' ')

# Calculate the starting and ending offsets of the Total column.
# Each output column is 20 characters wide in ifstat's "-w" mode
COLS_PER_IFACE=20
START=$(( IFCOUNT * COLS_PER_IFACE))
END=$(( IFCOUNT * COLS_PER_IFACE + COLS_PER_IFACE))

# Run ifstat until killed
# Ask for the Total column with -T
# Suppress periodic header with -n
# Fixed width output with -w
ifstat -n -w -T 0.1 $((10 * ($experimentation_time - 1) )) | cut -c${START}-${END}  > file

sed -r 's/[[:blank:]]+/,/g' file > fileBIS
sed  's/^.//' fileBIS > file2
num="$(wc -l file2 | cut -d " " -f1)"
echo $num
echo $(($num - 2 ))
tail file2 -n $(( $num - 2 )) > file3


echo "time_ms,INPUT_kB/s,OUTPUT_kB/s"> $filename  
lineno=0
while read dataline
do
   echo "$lineno,$dataline"
   lineno=$(($lineno+100))
done < ./file3 >> $filename

rm file file2 file3 fileBIS
