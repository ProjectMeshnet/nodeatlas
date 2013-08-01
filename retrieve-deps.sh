#!/bin/sh
#
## retrieve-deps.sh
#
# This script is included as part of NodeAtlas for retrieving static
# dependencies from content distribution networks (CDNs) or whatever
# permanent home they might have. The purpose of this is to keep them
# out of the Git repository, as per GitHub's recommendations on
# [repository size].
#
# When this script is run, a download tool will be chosen, which will
# default to either `curl` or `wget`, (in that order), depending on
# what is installed. If neither is installed, it will fail, unless a
# preferred command is specified.
#
# Dependencies should be listed in a file called `deps.txt` in the
# following format, one per line.
#
#     path/to/output/file.ext http://example.com/downloads/file.ext
#
# If a dependency can not be retrieved, its intended filename will be
# given on `stderr` with a warning that it could not be downloaded,
# and the script will continue. When it exits, however, and at least
# one dependency could not be retrieved, it will exit with status 3.
#
# If there is any other sort of error, it will exit with status 1.
#
# [repository size]: http://git.io/2uKFOw
#

#############
# Variables #
#############

DOWNLOADER=

#############
# Constants #
#############

DEFAULT_ASSETSDIR='res/web/assets'
DEFAULT_DEPFILE='depslist'
DEFAULT_DOWNLOADERS='curl -s -o,wget -qO'

COLOR_RED="\033[1;31m"
COLOR_DEFAULT="\033[0m"

#########################
# On to the actual code #
#########################

# Test if the dependency list is available and readable.
if [ ! -f "$DEFAULT_DEPFILE" ]; then
	echo "$DEFAULT_DEPFILE is not a regular file - exiting"
	exit 1
fi

# Select the first appropriate downloader, but only if DOWNLOADER is
# not specified.
if [ -z "$DOWNLOADER" ]; then
	# Set the IFS so that we only split the list on commas, but make
	# sure we can restore the old IFS. Note that DEFAULT_DOWNLOADERS
	# is not quoted in the for loop declaration.
	OLDIFS="$IFS"
	IFS=","
	for command in $DEFAULT_DOWNLOADERS; do
		# Check if the first word of the given download command is on
		# the system path. If so, choose it and break from the loop.
		which "$(echo $command | cut -d\  -f1)" > /dev/null
		if [ "$?" -eq 0 ]; then
			echo "Using $command as preferred downloader"
			DOWNLOADER="$command"
			break
		fi
	done

	# Restore the old IFS so that we don't make hard to find problems
	# later.
	IFS="$OLDIFS"

	# If we reach this point, and the DOWNLOADER is still not set, we
	# couldn't find an appropriate downloader, and we should quit.
	if [ -z "$DOWNLOADER" ]; then
		echo "Could not find an appropriate downloader"
		echo "Please specify one in $0"
		exit 1
	fi
fi

RETRIEVE_ALL=0

# Once the downloader has been selected, loop through the file line by
# line.
i=0
successes=0
while read outfile url; do
	i=$(($i + 1))

	# If either field is missing, explain that it is misformatted and
	# denote the failure, but continue.
	if [ -z "$outfile" -o -z "$url" ]; then
		echo " $COLOR_RED MISFORMATTED$COLOR_DEFAULT (line $i)"
		RETRIEVE_ALL=1
		continue
	fi

	# Report the status as we go.
	echo -n "  $outfile..."

	# The downloader should function such that '$DOWNLOADER
	# $DEFAULT_ASSETSDIR/$outfile $url' will retrieve a file at $url
	# and place it in the $DEFAULT_ASSETSDIR at $outfile.
	$DOWNLOADER "$DEFAULT_ASSETSDIR/$outfile" "$url" 1>&2

	# Check the exit status from the $DOWNLOADER, and make sure it's
	# zero. If it's not, denote the failure and continue.
	if [ "$?" -ne 0 ]; then
		echo " $COLOR_RED FAILED$COLOR_DEFAULT (line $i)"
		RETRIEVE_ALL=1
		continue
	fi
	
	# If all is well, state the success.
	echo " downloaded"
	successes=$(($successes + 1))
done < "$DEFAULT_DEPFILE"

# Report the status at the end, and exit with an appropriate code.
if [ "$RETRIEVE_ALL" -eq 0 ]; then
	echo "All dependencies downloaded: $i"
	exit 0
else
	echo "Dependencies downloaded: $successes (of $i)"
	echo "Errors were encountered during retrieval"
	exit 3
fi
