#!/bin/sh

# Enable debugging if variable is set
test $DEBUG_SHELL -eq 0 || set -o xtrace

cd "$(dirname "$0")"

timeout -s 9 $TIMEOUT /bin/sh tests.sh
code=$?
if [ $code -eq $((128+9)) ]
then
	echo "Execution time has exceed the timeout (TIMEOUT=$TIMEOUT)"
	exit 1
elif [ $code -ne 0 ]
then
	exit 1
fi
