#!/bin/bash

FLAGS="-i --rm"
NETWORK=tp0_testing_net
MSG="HolaServer"

RES=$(docker run $FLAGS --network=$NETWORK busybox sh -c "echo $MSG | nc server 12345")

if [ "$RES" == "$MSG" ]; then
  echo "action: test_echo_server | result: success"
else
  echo "action: test_echo_server | result: fail"
fi