#!/bin/bash

# we need to figure out the IP of the public interface here
IP=`ip addr show eth0 | grep "inet\b" | awk '{print $2}' | cut -d/ -f1`

nohup /home/coap/coap-server -address "${IP}" -password "connect:anything" -dTLS &
nohup /home/coap/coap-server -address "${IP}" &
sleep infinity