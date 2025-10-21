import os
from pathlib import PurePath

from app.util import load_env_files

PY_ROOT = str(PurePath(__file__).parent.parent)

load_env_files()


class BaseConfig:
    # App settings
    POSTS_PER_PAGE = os.environ.get('POSTS_PER_PAGE', 30)
    # STATIC_URL = os.environ.get('STATIC_URL', '/static')
    # STATIC_PATH = os.environ.get('STATIC_PATH', os.path.join(PY_ROOT, 'static'))

    # Server settings
    HTTP_IP = os.environ.get('HTTP_IP', '127.0.0.1')
    HTTP_PORT = int(os.environ.get('HTTP_PORT', 8000))
    MAX_CONTENT_LENGTH = int(os.environ.get('HTTP_MAX_BODY_SIZE', 10 * 1024 * 1024))

    # CORS settings
    CORS_ALLOWED_ORIGINS = os.environ.get('CORS_ALLOWED_ORIGINS', '*')
    CORS_ALLOWED_METHODS = os.environ.get(
        'CORS_ALLOWED_METHODS', 'GET, POST, PUT, DELETE, OPTIONS'
    )
    CORS_ALLOWED_HEADERS = os.environ.get(
        'CORS_ALLOWED_HEADERS', 'Content-Type, Authorization'
    )
    # TODO: allow credentials only when specific origins are set
    CORS_ALLOW_CREDENTIALS = (
        os.environ.get('CORS_ALLOW_CREDENTIALS', 'true').lower() == 'true'
    )
    CORS_MAX_AGE = int(os.environ.get('CORS_MAX_AGE', 86400))

    # Uploads settings
    UPLOAD_PATH = os.environ.get('UPLOAD_PATH', os.path.join(PY_ROOT, 'uploads'))
    UPLOAD_URL = os.environ.get('UPLOAD_URL', '/uploads')
    UPLOAD_IMAGE_FORMATS = ['png', 'jpg', 'jpeg', 'gif', 'webp']
    UPLOAD_THUMB_WIDTH = os.environ.get('UPLOAD_THUMB_WIDTH', 200)

    # Logging
    LOG_TYPE = os.environ.get('LOG_TYPE', 'console')  # 'console' or 'file'
    LOG_LEVEL = os.environ.get('LOG_LEVEL', 'INFO')
    LOG_FILE = os.environ.get('LOG_FILE', os.path.join(PY_ROOT, 'logs', 'app.log'))
    LOG_MAX_BYTES = os.environ.get('LOG_MAX_BYTES', 10 * 1024 * 1024)
    LOG_MAX_BACKUPS = os.environ.get('LOG_MAX_BACKUPS', 10)

    # Redis
    # REDIS = {
    #     'host': 'localhost',
    #     'port': 6379,
    #     'decode_responses': True,
    # }
    # Redis
    REDIS_URL = os.environ.get('REDIS_URL', 'redis://localhost:6379/0')

    # Misc
    ABOUT_URL = os.environ.get('ABOUT_URL')


class DevelopmentConfig(BaseConfig):
    # display query for debug
    # SQLALCHEMY_ECHO = True
    SQLALCHEMY_TRACK_MODIFICATIONS = False
    SQLALCHEMY_DATABASE_URI = os.environ.get('DATABASE_URL', 'sqlite:///app-dev.db')


class TestConfig(BaseConfig):
    SQLALCHEMY_DATABASE_URI = os.environ.get('DATABASE_URL', 'sqlite:///:memory:')

    # REDIS = {
    #     'host': 'localhost',
    #     'port': 6379,
    #     'db': 15,
    #     'decode_responses': True,
    # }


class ProductionConfig(BaseConfig):
    # sqlite
    SQLALCHEMY_TRACK_MODIFICATIONS = False
    SQLALCHEMY_DATABASE_URI = os.environ.get('DATABASE_URL', 'sqlite:///app.db')

    # logging
    LOG_TYPE = 'file'
    LOG_LEVEL = "WARNING"


config = ProductionConfig
env = os.environ.get('FLASK_ENV')

if env == 'development':
    config = DevelopmentConfig
elif env == 'test':
    config = TestConfig
else:
    config = ProductionConfig
