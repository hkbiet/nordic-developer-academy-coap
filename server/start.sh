#!/bin/bash

FILE=/run/secrets/STORAGE_CONNECTION_STRING     
if [ -f $FILE ]; then
export STORAGE_CONNECTION_STRING=$(cat /run/secrets/STORAGE_CONNECTION_STRING)
fi 

# we need to figure out the IP of the public interface here
IP=`ip addr show eth0 | grep "inet\b" | awk '{print $2}' | cut -d/ -f1`

nohup /home/coap/coap-server -address "${IP}" -password "connect:anything" -dTLS &
nohup /home/coap/coap-server -address "${IP}" &
sleep infinity