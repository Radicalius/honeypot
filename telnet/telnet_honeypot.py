import socket, sys, os, re
from socket import *

s = socket(AF_INET, SOCK_STREAM)
s.setsockopt(SOL_SOCKET, SO_REUSEADDR, 1)
s.setblocking(False)
s.bind(('0.0.0.0', 23))
s.listen(SOMAXCONN)

pid = os.fork()

if pid > 0:
    s.close()
    sys.exit(0)

log = open('/var/log/telnet_honeypot.log', 'a')

import logging
from twisted.internet import protocol, reactor, endpoints

os.setuid(1000)

logging.basicConfig(level=logging.DEBUG, stream=log, format='[%(asctime)s] %(message)s')

prompt = b'/ # '
proc_mounts = '''
proc /proc proc rw,nosuid,nodev,noexec,relatime 0 0
tmpfs /dev tmpfs rw,nosuid,size=65536k,mode=755,inode64 0 0
devpts /dev/pts devpts rw,nosuid,noexec,relatime,gid=5,mode=620,ptmxmode=666 0 0
sysfs /sys sysfs ro,nosuid,nodev,noexec,relatime 0 0
tmpfs /sys/fs/cgroup tmpfs rw,nosuid,nodev,noexec,relatime,mode=755,inode64 0 0
cgroup /sys/fs/cgroup/systemd cgroup ro,nosuid,nodev,noexec,relatime,xattr,name=systemd 0 0
cgroup /sys/fs/cgroup/pids cgroup ro,nosuid,nodev,noexec,relatime,pids 0 0
cgroup /sys/fs/cgroup/cpuset cgroup ro,nosuid,nodev,noexec,relatime,cpuset 0 0
cgroup /sys/fs/cgroup/cpu,cpuacct cgroup ro,nosuid,nodev,noexec,relatime,cpu,cpuacct 0 0
cgroup /sys/fs/cgroup/hugetlb cgroup ro,nosuid,nodev,noexec,relatime,hugetlb 0 0
cgroup /sys/fs/cgroup/freezer cgroup ro,nosuid,nodev,noexec,relatime,freezer 0 0
cgroup /sys/fs/cgroup/perf_event cgroup ro,nosuid,nodev,noexec,relatime,perf_event 0 0
cgroup /sys/fs/cgroup/net_cls,net_prio cgroup ro,nosuid,nodev,noexec,relatime,net_cls,net_prio 0 0
cgroup /sys/fs/cgroup/misc cgroup ro,nosuid,nodev,noexec,relatime,misc 0 0
cgroup /sys/fs/cgroup/rdma cgroup ro,nosuid,nodev,noexec,relatime,rdma 0 0
cgroup /sys/fs/cgroup/devices cgroup ro,nosuid,nodev,noexec,relatime,devices 0 0
cgroup /sys/fs/cgroup/blkio cgroup ro,nosuid,nodev,noexec,relatime,blkio 0 0
cgroup /sys/fs/cgroup/memory cgroup ro,nosuid,nodev,noexec,relatime,memory 0 0
mqueue /dev/mqueue mqueue rw,nosuid,nodev,noexec,relatime 0 0
shm /dev/shm tmpfs rw,nosuid,nodev,noexec,relatime,size=65536k,inode64 0 0
/dev/sdb2 /etc/resolv.conf ext4 rw,relatime,errors=remount-ro 0 0
/dev/sdb2 /etc/hostname ext4 rw,relatime,errors=remount-ro 0 0
/dev/sdb2 /etc/hosts ext4 rw,relatime,errors=remount-ro 0 0
devpts /dev/console devpts rw,nosuid,noexec,relatime,gid=5,mode=620,ptmxmode=666 0 0
proc /proc/bus proc ro,nosuid,nodev,noexec,relatime 0 0
proc /proc/fs proc ro,nosuid,nodev,noexec,relatime 0 0
proc /proc/irq proc ro,nosuid,nodev,noexec,relatime 0 0
proc /proc/sys proc ro,nosuid,nodev,noexec,relatime 0 0
proc /proc/sysrq-trigger proc ro,nosuid,nodev,noexec,relatime 0 0
tmpfs /proc/asound tmpfs ro,relatime,inode64 0 0
tmpfs /proc/acpi tmpfs ro,relatime,inode64 0 0
tmpfs /proc/kcore tmpfs rw,nosuid,size=65536k,mode=755,inode64 0 0
tmpfs /proc/keys tmpfs rw,nosuid,size=65536k,mode=755,inode64 0 0
tmpfs /proc/timer_list tmpfs rw,nosuid,size=65536k,mode=755,inode64 0 0
tmpfs /proc/scsi tmpfs ro,relatime,inode64 0 0
tmpfs /sys/firmware tmpfs ro,relatime,inode64 0 0
'''.strip().encode()

class TelnetProtocol(protocol.Protocol):

    def __init__(self):
        self.mode = 'username'
        self.username = ''

    def connectionMade(self):
        self.transport.write(b'Username: ')

    def dataReceived(self, data):

        data = data.strip()

        if data in [b'exit', b'quit']:
            self.transport.loseConnection()

        if self.mode == 'username':
            self.username = data
            self.transport.write(b'Password: ')
            self.mode = 'password'
        elif self.mode == 'password':
            self.transport.write(prompt)
            logging.info(f'{self.transport.getPeer().host}: login ({self.username.decode()}. {data.decode()})')
            self.mode = 'commands'
        elif self.mode == 'commands':
            logging.info(f'{self.transport.getPeer().host}: > {data.decode()}')

            for command in data.decode().split(';'):
                try:
                    command = command.strip()
                    if re.match('.*[A-Z]{5}.*', command):
                        token = re.findall('[A-Z]{5}', command)[0]
                        self.transport.write(f'{token}: applet not found\r\n'.encode())
                    if '/proc/mounts' in command:
                        self.transport.write(proc_mounts)
                        self.transport.write('\r\n'.encode())
                except:
                    logging.exception(f'Error when parsing command {data.decode()}: ')

            self.transport.write(prompt)

class TelnetProtocolFactory(protocol.Factory):
    def buildProtocol(self, addr):
        return TelnetProtocol()

reactor.adoptStreamPort(s.fileno(), AF_INET, TelnetProtocolFactory())
reactor.run()
                  