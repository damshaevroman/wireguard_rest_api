#!/bin/bash
cp wireguard-rest /usr/bin/wireguard-rest
cp wireguard_api.cfg /etc/wireguard_api.cfg 
cp wireguard-rest.service /lib/systemd/system/wireguard-rest.service
systemctl enable wireguard-rest.service
systemctl start wireguard-rest.service
