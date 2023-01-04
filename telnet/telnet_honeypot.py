import socket, sys, os
from socket import *

s = socket(AF_INET, SOCK_STREAM)
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

intro_text = r"""
/\ \/\ \/\ \                    /\ \__           
\ \ \ \ \ \ \____  __  __    ___\ \ ,_\  __  __  
 \ \ \ \ \ \ '__`\/\ \/\ \ /' _ `\ \ \/ /\ \/\ \ 
  \ \ \_\ \ \ \L\ \ \ \_\ \/\ \/\ \ \ \_\ \ \_\ \
   \ \_____\ \_,__/\ \____/\ \_\ \_\ \__\\ \____/
    \/_____/\/___/  \/___/  \/_/\/_/\/__/ \/___/ 
                                                 
Welcome to Ubuntu 20.04.2 LTS (GNU/Linux 5.4.0-1077 aarch64)

 * Documentation:  https://help.ubuntu.com
 * Management:     https://landscape.canonical.com
 * Support:        https://ubuntu.com/advantage

  System information

  System load:                      0.28
  Usage of /:                       72.0% of 29.03GB
  Memory usage:                     48%
  Swap usage:                       0%
  Temperature:                      52.1 C
  Processes:                        177
  Users logged in:                  0
  IPv4 address for eth0:            10.20.35.11

117 updates can be installed immediately.
0 of these updates are security updates.
To see these additional updates run: apt list --upgradable


Last login: Tue Dec 13 04:02:22 2022

"""

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
            self.transport.write(intro_text.encode('ascii'))
            self.transport.write(self.username + b'@ubuntu:~$ ')
            logging.info(f'{self.username.decode()}@{self.transport.getPeer().host}: ? {data.decode()}')
            self.mode = 'commands'
        elif self.mode == 'commands':
            logging.info(f'{self.username.decode()}@{self.transport.getPeer().host}: > {data.decode()}')
            self.transport.write(self.username + b'@ubuntu:~$ ')

class TelnetProtocolFactory(protocol.Factory):
    def buildProtocol(self, addr):
        return TelnetProtocol()

reactor.adoptStreamPort(s.fileno(), AF_INET, TelnetProtocolFactory())
reactor.run()
                  