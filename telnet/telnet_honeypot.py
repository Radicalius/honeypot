import socket, sys, os, re, base64
from socket import *

from elf_header import random_elf_header

wd = os.environ.get('WORKING_DIRECTORY') or os.getcwd()
port = 23 if os.getuid() == 0 else 2300

s = socket(AF_INET, SOCK_STREAM)
s.setsockopt(SOL_SOCKET, SO_REUSEADDR, 1)
s.setblocking(False)
s.bind(('0.0.0.0', port))
s.listen(SOMAXCONN)

pid = os.fork()

if pid > 0:
    s.close()
    sys.exit(0)

import logging
from twisted.internet import protocol, reactor, endpoints

os.setgid(1000)
os.setuid(1000)

log = open(f'{wd}/telnet.log', 'a')
logging.basicConfig(level=logging.DEBUG, stream=log, format='[%(asctime)s] %(message)s')

prompt = b'/ # '
proc_mounts = base64.b64decode(open(f'{wd}/data/proc_mounts.b64').read())
wget_help = base64.b64decode(open(f'{wd}/data/wget_help.b64').read())
# busybox_binary = base64.b64decode(open(f'{wd}/data/busybox.b64').read())
elf_template = base64.b64decode(open(f'{wd}/data/echo.b64').read())

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
            logging.info(f'{self.transport.getPeer().host}: login ({self.username.decode()}, {data.decode()})')
            self.mode = 'commands'
        elif self.mode == 'commands':
            try:
                data = data.decode()
            except:
                logging.warning(f'{self.transport.getPeer().host}: Undecodable command received: {data}')
                return

            logging.info(f'{self.transport.getPeer().host}: > {data}')

            for command in data.split(';'):
                try:
                    command = command.strip()
                    if re.match('.*[A-Z]{5,6}.*', command):
                        token = re.findall('[A-Z]{5,6}', command)[0]
                        self.transport.write(f'{token}: applet not found\r\n'.encode())
                    if '/proc/mounts' in command:
                        self.transport.write(proc_mounts)
                        self.transport.write('\r\n'.encode())
                    if re.match(r'.*echo -n?e \'(\\x[0-9a-f]{2})+\'', command):
                        for digit in re.findall(r'\\x[0-9a-f]{2}', command):
                            digit = digit.replace('\\x', '')
                            self.transport.write(chr(int(digit, 16)).encode())
                        self.transport.write(b'\r\n')
                    if command == 'tftp':
                        self.transport.write('tftp: applet not found\r\n'.encode())
                    if command == 'wget':
                        self.transport.write(wget_help)
                        self.transport.write('\r\n'.encode())
                    if 'dd' in command and 'if=.s' in command and 'cat .s' in command:
                        elf_header, endian, arch = random_elf_header(elf_template)
                        logging.info(f'{self.transport.getPeer().host}: masquerading as ({endian}, {arch})')
                        self.transport.write(elf_header)
                        self.transport.write('\r\n'.encode())
                except:
                    logging.exception(f'Error when parsing command {data}: ')

            self.transport.write(prompt)

class TelnetProtocolFactory(protocol.Factory):
    def buildProtocol(self, addr):
        return TelnetProtocol()

reactor.adoptStreamPort(s.fileno(), AF_INET, TelnetProtocolFactory())
reactor.run()
                  