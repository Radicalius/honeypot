#!/bin/bash

user=${whoami}

sudo mkdir -p /usr/local/lib/telnet_honeypot
sudo cp telnet_honeypot.py /usr/local/lib/telnet_honeypot

sudo cp telnet_honeypot.service /etc/systemd/system/
sudo systemctl daemon-reload
sudo systemctl enable telnet_honeypot
sudo systemctl start telnet_honeypot