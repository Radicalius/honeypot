import flask, os, logging, base64
from flask import Flask, request, Response
app = Flask(__name__)

wd = os.environ.get('WORKING_DIRECTORY') or os.getcwd()

log = open(f'{wd}/http.log', 'a')
logging.basicConfig(level=logging.DEBUG, stream=log, format='[%(asctime)s] %(message)s')

laravel6_env_data = base64.b64decode(open(f'{wd}/data/laravel6.env.b64').read())

@app.route('/', defaults={'path': ''})
@app.route('/<path:path>')
def catch_all(path):
    logging.info(f'{request.headers.get("X-Remote-IP")} - {request.method} {request.path}\n{request.headers}\n{request.data}')
    
    if request.path == '/.env':
        resp = Response(laravel6_env_data)
        resp.headers['Content-Type'] = 'text/plain'
        return resp

    return '404 File not found.', 404
