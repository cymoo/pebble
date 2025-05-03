"""
Application factory
~~~~~~~~~~~~~~~~~~~
"""

import os
from dataclasses import is_dataclass

from dotenv import load_dotenv
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

    db.init_app(app)
    db.app = app

    rd = Redis(**app.config['REDIS'])
    app.rd = rd

    app.searcher = FullTextSearch(
        rd, app.config['PARTIAL_MATCH'], app.config['SEARCH_MAX_RESULTS']
    )


def register_blueprints(app: Flask) -> None:
    from .api import api
    from .view import view

    app.register_blueprint(api, url_prefix='/api')
    app.register_blueprint(view, url_prefix='/shared')


def register_file_uploads(app: Flask) -> None:
    upload_folder = app.config['UPLOAD_FOLDER']
    if not os.path.exists(upload_folder):
        os.mkdir(upload_folder)

    @app.route('/uploads/<path:filename>')
    def uploaded_file(filename):
        return send_from_directory(upload_folder, filename)


def configure_cors(app: Flask) -> None:
    @app.after_request
    def set_cors_headers(res: Response) -> Response:
        config = app.config
        headers = res.headers
        headers['Access-Control-Allow-Origin'] = config['ACCESS_CONTROL_ALLOW_ORIGIN']
        headers['Access-Control-Allow-Methods'] = config['ACCESS_CONTROL_ALLOW_METHODS']
        headers['Access-Control-Allow-Headers'] = config['ACCESS_CONTROL_ALLOW_HEADERS']
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
