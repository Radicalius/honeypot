#!/bin/bash

python3 -c 'import os;open("telnet_honeypot.service", "w").write(open("telnet_honeypot.service.tmp", "r").read().replace("{{wd}}", os.getcwd()))'

if [ ! -f /etc/systemd/system/telnet_honeypot.service ]; then
    sudo ln -s "$(pwd)/telnet_honeypot.service" /etc/systemd/system/
fi

sudo systemctl daemon-reload
sudo systemctl enable telnet_honeypot
sudo systemctl restart telnet_honeypot