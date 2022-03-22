#!/bin/bash

#
# Make sure the code is formatted well.
#

badly_formatted_files="$(goimports -l ./)"

if [ -n "$badly_formatted_files" ]; then
	echo
	echo "     The following files are formatted badly:"
	echo

	echo "$badly_formatted_files" | while read line; do
		echo "         $line"
	done

	exit 1
fi
