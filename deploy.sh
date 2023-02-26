#!/bin/bash

. secrets.env

command() {
    ssh -o StrictHostKeychecking=no -i -p $PORT "$PEM_FILE" "$USERNAME@$HOST_IP" $1
}

copy() {
    scp -o StrictHostKeychecking=no -r -i -p $PORT "$PEM_FILE" "$1" "$USERNAME@$HOST_IP:$2"
}

for dir in `ls ./services`; do
    cd "services/$dir" && \
    CGO_ENABLED=0 go build -a -installsuffix cgo && \
    cd "../.." && \
    command "sudo rm -rf /usr/local/bin/${dir}" && \
    command "rm -rf ~/staging/${dir}" && \
    copy "services/${dir}" "~/staging" && \
    copy "appsettings.env" "~/staging/${dir}" && \
    command "sudo mv ~/staging/${dir} /usr/local/bin" && \
    command "sudo chmod a+x /usr/local/bin/${dir}/${dir}"
    command "sudo setcap cap_net_bind_service=+ep /usr/local/bin/${dir}/${dir}"

    copy "openrc/basic.tmp.sh" "~/staging/${dir}.sh" && \
    command "sudo mv staging/${dir}.sh /etc/init.d/${dir}" && \
    command "sudo chmod a+rx /etc/init.d/${dir}" && \
    command "sudo rc-update add ${dir} default" && \
    command "sudo rc-service ${dir} restart"
done