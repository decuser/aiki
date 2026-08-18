#!/usr/bin/env bash
set -eu

TITLE=$1
LOGFILE=$2
COORD_PID=$3

printf '\033]0;%s\007' "$TITLE"
printf '%s\n' "$TITLE"
printf '%s\n' "----------------------------------------"

# The log file is created by showcase.sh before this viewer is launched.
# --pid makes the viewer terminate when the real coordinator terminates.
tail --pid="$COORD_PID" -n +1 -f "$LOGFILE" || true

printf '\n%s stopped\n' "$TITLE"
sleep 1
