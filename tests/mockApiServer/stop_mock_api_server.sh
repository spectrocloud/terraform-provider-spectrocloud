#!/bin/bash

# Find the process ID of the running mockApiserver
PID=$(pgrep -f MockBuild)

if [ -z "$PID" ]; then
  echo "MockAPIServer is not running."
else
  # Kill the process
  kill $PID
  echo "MockAPIServer (PID: $PID) has been stopped."
fi
[ -f "./MockBuild" ] && rm -f "./MockBuild"