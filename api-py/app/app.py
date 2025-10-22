"""
Application factory
~~~~~~~~~~~~~~~~~~~
"""

import os
from dataclasses import is_dataclass

from flask import Flask, Response, send_from_directory, jsonify, request, make_response
from flask_sqlalchemy import SQLAlchemy
from flask.json.provider import JSONProvider
from orjson import orjson


def create_app(config) -> Flask:
    """Create Flask application instance"""

    from .exception import register_error_handlers
    from .task import huey

    app = Flask(
        'app',
        static_folder='../static',
        template_folder='../templates',
    )

    make_response_of_dataclass(app)

    app.json = ORJSONProvider(app)

    app.config.from_object(config)

    init_extension(app)
    register_blueprints(app)
    register_file_uploads(app)
    register_error_handlers(app)
    register_cors_handlers(app)

    # Run periodic tasks
    huey.start()

    return app


def init_extension(app: Flask) -> None:
    from .extension import init_extension, db
    from sqlalchemy import text

    init_extension(app)

    # enable foreign key constraint and WAL mode for SQLite
    with app.app_context():
        db.session.execute(text("PRAGMA foreign_keys=ON"))
        db.session.execute(text("PRAGMA journal_mode=WAL"))

    # Run migrations if enabled
    if app.config['AUTO_MIGRATE']:
        run_migration(app, db)


def register_blueprints(app: Flask) -> None:
    from .handler.api import api
    from .handler.page import page

    app.register_blueprint(api, url_prefix='/api')
    app.register_blueprint(page, url_prefix='/shared')


def register_file_uploads(app: Flask) -> None:
    upload_path = app.config['UPLOAD_PATH']
    upload_url = app.config['UPLOAD_URL']
    if not os.path.exists(upload_path):
        os.mkdir(upload_path)

    @app.route(f'{upload_url}/<path:filename>')
    def uploaded_file(filename):
        abs_path = os.path.abspath(upload_path)
        return send_from_directory(abs_path, filename)


def register_cors_handlers(app: Flask) -> None:
    """Register CORS handling functions"""

    # Get CORS settings from config
    config = app.config
    allowed_origins = config['CORS_ALLOWED_ORIGINS']
    allowed_methods = config['CORS_ALLOWED_METHODS']
    allowed_headers = config['CORS_ALLOWED_HEADERS']
    allow_credentials = config['CORS_ALLOW_CREDENTIALS']
    max_age = config['CORS_MAX_AGE']

    # Handle allowed origins
    origins_list = (
        [o.strip() for o in allowed_origins.split(',')]
        if allowed_origins != '*'
        else ['*']
    )

    def is_origin_allowed(origin: str) -> bool:
        """check if the origin is allowed"""
        if '*' in origins_list:
            return True
        return origin in origins_list

    def set_cors_headers(response: Response, origin: str) -> Response:
        """set CORS headers to the response"""
        # Set Access-Control-Allow-Origin
        if '*' in origins_list:
            response.headers['Access-Control-Allow-Origin'] = '*'
        else:
            response.headers['Access-Control-Allow-Origin'] = origin
            # When allowing specific domains, add Vary header to support caching
            response.headers['Vary'] = 'Origin'

        # Set allowed methods
        response.headers['Access-Control-Allow-Methods'] = allowed_methods

        # Set allowed headers
        response.headers['Access-Control-Allow-Headers'] = allowed_headers

        # Set Access-Control-Allow-Credentials
        if allow_credentials:
            response.headers['Access-Control-Allow-Credentials'] = 'true'

        # Set Access-Control-Max-Age
        response.headers['Access-Control-Max-Age'] = max_age

        return response

    @app.before_request
    def handle_preflight() -> Response | None:
        "Handle CORS preflight requests"
        if request.method == 'OPTIONS':
            origin = request.headers.get('Origin')

            if origin and is_origin_allowed(origin):
                response = make_response('', 204)
                set_cors_headers(response, origin)
                return response
            # If origin is not allowed, return 403
            elif origin:
                return make_response('', 403)

    @app.after_request
    def add_cors_headers(response) -> Response:
        """Add CORS headers to all responses"""
        origin = request.headers.get('Origin')

        # If no Origin header, no need to handle CORS (same-origin request)
        if not origin:
            return response

        # Check if origin is allowed
        if is_origin_allowed(origin):
            set_cors_headers(response, origin)

        return response


def make_response_of_dataclass(app):
    original_make_response = app.make_response

    def new_make_response(rv):
        if is_dataclass(rv):
            rv = jsonify(rv)
        return original_make_response(rv)

    app.make_response = new_make_response


def run_migration(app: Flask, db: SQLAlchemy) -> None:
    from flask_migrate import Migrate, upgrade

    _ = Migrate(app, db)
    logger = app.logger

    with app.app_context():
        if not os.path.exists('migrations'):
            logger.info("Migrations folder not found")
            logger.info("Run the following command to initialize migrations:")
            logger.info("    flask db init")
            logger.info("    flask db migrate -m 'Initial migration'")
            logger.info("    flask db upgrade")
            logger.info("Now creating the database tables directly.")
            db.create_all()
        else:
            try:
                upgrade()
                logger.info("Database migrated to latest version")
            except Exception as e:
                logger.error(f"Database migration failed: {e}")
                logger.info("Creating database tables directly.")
                db.create_all()


class ORJSONProvider(JSONProvider):
    """Custom JSON provider using orjson for Flask applications."""

    def __init__(self, *args, **kwargs):
        self.options = kwargs
        super().__init__(*args, **kwargs)

    def loads(self, s, **kwargs):
        return orjson.loads(s)

    def dumps(self, obj, **kwargs):
        # Decode back to str, as orjson returns bytes
        return orjson.dumps(obj, option=orjson.OPT_NON_STR_KEYS).decode('utf-8')
