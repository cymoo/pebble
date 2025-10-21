"""
Application factory
~~~~~~~~~~~~~~~~~~~
"""

import os
from dataclasses import is_dataclass

from flask import Flask, send_from_directory, Response, jsonify
from flask.json.provider import JSONProvider
from orjson import orjson

from .exception import register_error_handlers
from .task import huey
from .logger import setup_logger


def create_app(config) -> Flask:
    setup_logger(config)

    app = Flask('app', static_folder='../static', template_folder='../templates')

    make_response_of_dataclass(app)

    app.json = ORJSONProvider(app)

    app.config.from_object(config)

    register_db(app)
    register_blueprints(app)
    register_file_uploads(app)
    register_error_handlers(app)
    configure_cors(app)

    # Run periodic tasks
    huey.start()

    return app


def register_db(app: Flask) -> None:
    from redis import Redis
    from .model import db
    from .lib.search import FullTextSearch
    from sqlalchemy import text

    db.init_app(app)
    db.app = app

    # enable foreign key constraint and WAL mode for SQLite
    with app.app_context():
        db.session.execute(text("PRAGMA foreign_keys=ON"))
        db.session.execute(text("PRAGMA journal_mode=WAL"))

    rd = Redis.from_url(app.config['REDIS_URL'], decode_responses=True)
    app.rd = rd

    app.fts = FullTextSearch(rd, 'fts:')


def register_blueprints(app: Flask) -> None:
    from .api import api
    from .view import view

    app.register_blueprint(api, url_prefix='/api')
    app.register_blueprint(view, url_prefix='/shared')


def register_file_uploads(app: Flask) -> None:
    upload_path = app.config['UPLOAD_PATH']
    upload_url = app.config['UPLOAD_URL']
    if not os.path.exists(upload_path):
        os.mkdir(upload_path)

    @app.route(f'{upload_url}/<path:filename>')
    def uploaded_file(filename):
        print('xxx', upload_path, filename)
        abs_path = os.path.abspath(upload_path)
        return send_from_directory(abs_path, filename)


def configure_cors(app: Flask) -> None:
    @app.after_request
    def set_cors_headers(res: Response) -> Response:
        config = app.config
        headers = res.headers
        headers['Access-Control-Allow-Origin'] = config['CORS_ALLOWED_ORIGINS']
        headers['Access-Control-Allow-Methods'] = config['CORS_ALLOWED_METHODS']
        headers['Access-Control-Allow-Headers'] = config['CORS_ALLOWED_HEADERS']
        # TODO: allow credentials only when specific origins are set, and set max age
        return res


class ORJSONProvider(JSONProvider):
    def __init__(self, *args, **kwargs):
        self.options = kwargs
        super().__init__(*args, **kwargs)

    def loads(self, s, **kwargs):
        return orjson.loads(s)

    def dumps(self, obj, **kwargs):
        # decode back to str, as orjson returns bytes
        return orjson.dumps(obj, option=orjson.OPT_NON_STR_KEYS).decode('utf-8')


def make_response_of_dataclass(app):
    original_make_response = app.make_response

    def new_make_response(rv):
        if is_dataclass(rv):
            rv = jsonify(rv)
        return original_make_response(rv)

    app.make_response = new_make_response
