#!/bin/bash

# Find the process ID of the running mockApiserver
PID=$(pgrep -f MockAPIServer)

if [ -z "$PID" ]; then
  echo "MockAPIServer is not running."
else
  # Kill the process
  kill $PID
#  [ -f ./MockAPIServer ] && rm -f ./MockAPIServer
  echo "MockAPIServer (PID: $PID) has been stopped."
fi