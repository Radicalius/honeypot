import socket, sys, os, re, base64
from socket import *

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

from twisted.internet import protocol, reactor, endpoints

from log import Logger
from elf_header import random_elf_header

os.setgid(1000)
os.setuid(1000)

prompt = b'/ # '
proc_mounts = base64.b64decode(open(f'{wd}/data/proc_mounts.b64').read())
wget_help = base64.b64decode(open(f'{wd}/data/wget_help.b64').read())
# busybox_binary = base64.b64decode(open(f'{wd}/data/busybox.b64').read())
elf_template = base64.b64decode(open(f'{wd}/data/echo.b64').read())

class TelnetProtocol(protocol.Protocol):

    def __init__(self):
        self.mode = 'username'
        self.logger = Logger('telnet')

    def connectionMade(self):
        self.logger.bind(ip=self.transport.getPeer().host)
        self.transport.write(b'Username: ')

    def dataReceived(self, data):

        data = data.strip()

        try:
            data = data.decode()
        except:
            self.logger.warning(message='Undecodable command received', action='command', command=data)
            return

        if data in ['exit', 'quit']:
            self.transport.loseConnection()

        if self.mode == 'username':
            self.logger.bind(username=data)
            self.transport.write(b'Password: ')
            self.mode = 'password'
        elif self.mode == 'password':
            self.transport.write(prompt)
            self.logger.bind(password=data)
            self.logger.info(action='login')
            self.mode = 'commands'
        elif self.mode == 'commands':

            self.logger.info(action='command', command=data)

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
                        self.logger.bind(endianess=endian, architecture=arch)
                        self.logger.info(action='masquerade')
                        self.transport.write(elf_header)
                        self.transport.write('\r\n'.encode())
                except:
                    self.logger.exception(message='Error occurred while parsing command', action='command', command=data)

            self.transport.write(prompt)

class TelnetProtocolFactory(protocol.Factory):
    def buildProtocol(self, addr):
        return TelnetProtocol()

reactor.adoptStreamPort(s.fileno(), AF_INET, TelnetProtocolFactory())
reactor.run()
                  