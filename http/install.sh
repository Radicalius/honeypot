#!/bin/bash

python3 -c 'import os;open("http_honeypot.service", "w").write(open("http_honeypot.service.tmp", "r").read().replace("{{wd}}", os.getcwd()).replace("{{user}}", os.getlogin()))'

if [ ! -f /etc/systemd/system/http_honeypot.service ]; then
    sudo ln -s "$(pwd)/http_honeypot.service" /etc/systemd/system/
fi

sudo mkdir -p /var/log/honeypots

sudo systemctl daemon-reload
sudo systemctl enable telnet_honeypot
sudo systemctl restart telnet_honeypot

sudo rm /etc/nginx/sites-enabled
sudo ln -s "$(pwd)/nginx.conf" /etc/nginx/sites-enabled/nginx.conf
sudo systemctl reload nginx
