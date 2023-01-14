#!/bin/bash

python3 -c 'import os;open("ftp_honeypot.service", "w").write(open("ftp_honeypot.service.tmp", "r").read().replace("{{wd}}", os.getcwd()))'

if [ ! -f /etc/systemd/system/ftp_honeypot.service ]; then
    sudo ln -s "$(pwd)/ftp_honeypot.service" /etc/systemd/system/
fi

sudo mkdir -p /var/log/honeypots

sudo systemctl daemon-reload
sudo systemctl enable ftp_honeypot
sudo systemctl restart ftp_honeypot