#!/sbin/openrc-run

command="/usr/local/bin/${RC_SVCNAME}/${RC_SVCNAME}"
directory="/usr/local/bin/${RC_SVCNAME}"
command_user="honeypot:honeypot"
command_background=true
pidfile="/run/${RC_SVCNAME}.pid"