import flask, os, logging
from flask import Flask, request
app = Flask(__name__)

wd = os.environ.get('WORKING_DIRECTORY') or os.getcwd()

log = open(f'{wd}/http.log', 'a')
logging.basicConfig(level=logging.DEBUG, stream=log, format='[%(asctime)s] %(message)s')

@app.route('/', defaults={'path': ''})
@app.route('/<path:path>')
def catch_all(path):
    logging.info(f'{request.remote_addr} - {request.method} {request.path}\n{request.headers}\n{request.data}')
    return '404 File not found.', 404
