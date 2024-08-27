#!/bin/bash

# Find the process ID of the running mockApiserver
PID=$(pgrep -f MockBuild)

if [ -z "$PID" ]; then
  echo "MockAPIServer is not running."
else
  # Kill the process
  kill $PID
  kill -9 $(lsof -t -i :8080) $(lsof -t -i :8888)
  echo "MockAPIServer (PID: $PID) has been stopped."
fi