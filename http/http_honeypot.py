import flask, os, base64
from flask import Flask, request, Response

from log_ import Logger
logger = Logger("http")

app = Flask(__name__)

wd = os.environ.get('WORKING_DIRECTORY') or os.getcwd()

laravel6_env_data = base64.b64decode(open(f'{wd}/data/laravel6.env.b64').read())

@app.route('/', defaults={'path': ''})
@app.route('/<path:path>')
def catch_all(path):
    headers = {k: v for k,v in request.headers}
    logger.info(ip=request.headers.get("X-Remote-IP"), method=request.method, path=request.path, headers=headers, data=request.data.decode())
    
    if request.path == '/.env':
        resp = Response(laravel6_env_data)
        resp.headers['Content-Type'] = 'text/plain'
        return resp

    return '404 File not found.', 404
