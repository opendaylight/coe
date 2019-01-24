#!/bin/sh
PART1=$(echo "$@" | cut -c1-8)
PART2=$(echo "$@" | cut -c9-12)
PART3=$(echo "$@" | cut -c13-16)
PART4=$(echo "$@" | cut -c17-20)
PART5=$(echo "$@" | cut -c21-32)
echo "$PART1-$PART2-$PART3-$PART4-$PART5"
echo "$PART1"
