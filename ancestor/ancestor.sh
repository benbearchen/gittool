#!/bin/sh

cd $1

A=$2
B=$3
if [ "$A" = "" -o "$B" = "" ]
then
	echo ""
	echo "usage: sh ancestor.sh <repodir> <commit-new> <commit-old>"
	exit
fi

AHASH=`git merge-base $A $A`
if [ "$AHASH" = "" ]
then
	echo ""
	echo "invalid commit/branch $A"
	exit
fi

BHASH=`git merge-base $B $B`
if [ "$BHASH" = "" ]
then
	echo ""
	echo "invalid commit/branch $B"
	exit
fi

C=`git merge-base $A $B`
if [ "$BHASH" = "$C" ]
then
	echo "1"
else
	echo "0"
fi

