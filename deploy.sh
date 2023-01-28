#!/bin/bash

. secrets.env

command() {
    ssh -i "$PEM_FILE" "$USERNAME@$HOST_IP" $1
}

for dir in `ls ./modules`; do
    cd "modules/$dir" && \
    CGO_ENABLED=0 go build -a -installsuffix cgo && \
    cd "../.." && \
    command "sudo rm -rf /usr/local/bin/${dir}_honeypot" && \
    command "rm -rf ~/staging/${dir}" && \
    scp -r -i "$PEM_FILE" "modules/${dir}" "$USERNAME@$HOST_IP:~/staging" && \
    command "sudo mv ~/staging/${dir} /usr/local/bin" && \
    command "sudo mv /usr/local/bin/${dir} /usr/local/bin/${dir}_honeypot" && \
    command "sudo chmod a+x /usr/local/bin/${dir}_honeypot"

    scp -r -i "$PEM_FILE" "openrc/basic.tmp.sh" "$USERNAME@$HOST_IP:~/staging/${dir}_honeypot.sh" && \
    command "sudo mv staging/${dir}_honeypot.sh /etc/init.d/${dir}_honeypot" && \
    command "sudo chmod a+rx /etc/init.d/${dir}_honeypot" && \
    command "sudo rc-update add ${dir}_honeypot default" && \
    command "sudo rc-service ${dir}_honeypot restart"
done