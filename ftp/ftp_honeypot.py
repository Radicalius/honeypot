import socket, sys, os, re, base64
from socket import *

wd = os.environ.get('WORKING_DIRECTORY') or os.getcwd()
port = 21 if os.getuid() == 0 else 2100

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
from log_ import Logger

os.setgid(1000)
os.setuid(1000)

class FtpProtocol(protocol.Protocol):

    def __init__(self):
        self.logger = Logger('ftp')

    def connectionMade(self):
        self.logger.bind(ip=self.transport.getPeer().host)
        self.transport.write(b'220 Welcome to the DLP Test FTP Server\r\n')

    def dataReceived(self, data):
        try:
            data = data.decode().strip()
        except:
            self.logger.exception(message='Error parsing command.', action='command', command=str(data))

        self.logger.info(action='command', command=data)

        parts = data.split(' ')
        cmd = parts[0]
        
        if cmd == 'USER':
            self.logger.bind(username=parts[1])
            self.transport.write(b'331 Please specify the password.\r\n')
        elif cmd == 'PASS':
            self.logger.bind(password=parts[1])
            self.logger.info(action='login')
            self.transport.write(b'230 Login successful.\r\n')
        elif cmd == 'SYST':
            self.transport.write(b'215 UNIX Type: L8\r\n')
        elif cmd == 'CLOSE' or cmd == 'QUIT' or cmd == 'EXIT':
            self.transport.write(b'221 Goodbye.\r\n')
            self.transport.loseConnection()
        else:
            self.transport.write(b'500 Invalid command.\r\n')
        

class FtpProtocolFactory(protocol.Factory):
    def buildProtocol(self, addr):
        return FtpProtocol()

reactor.adoptStreamPort(s.fileno(), AF_INET, FtpProtocolFactory())
reactor.run()