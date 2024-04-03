#!/bin/bash

FILE=/run/secrets/STORAGE_CONNECTION_STRING     
if [ -f $FILE ]; then
export STORAGE_CONNECTION_STRING=$(cat /run/secrets/STORAGE_CONNECTION_STRING)
fi 

# Enumerate all IPv4 network interfaces and get their assigned IP addresses
ipv4_interfaces=$(ip -o -4 addr show | awk '{print $2}' | cut -d':' -f1)

for interface in $ipv4_interfaces; do
    IP=$(ip -o -4 addr show $interface | awk '{print $4}' | cut -d'/' -f1)
    echo "Interface: $interface, Assigned IP: $IP"

    nohup /home/coap/coap-server -address "${IP}" -password "connect:anything" -dTLS &
    nohup /home/coap/coap-server -address "${IP}" &
done

# Enumerate all IPv6 network interfaces and get their assigned IP addresses
ipv6_interfaces=$(ip -o -6 addr show | awk '{print $2}' | cut -d':' -f1)

for interface in $ipv6_interfaces; do
    IP=$(ip -o -6 addr show $interface | awk '{print $4}' | cut -d'/' -f1)
    echo "Interface: $interface, Assigned IPv6: $IP"

    nohup /home/coap/coap-server -address "${IP}" -udp6 -password "connect:anything" -dTLS &
    nohup /home/coap/coap-server -address "${IP}" -udp6 &
done

sleep infinity