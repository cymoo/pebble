import os
from pathlib import PurePath

from app.util import load_env_files

PY_ROOT = str(PurePath(__file__).parent.parent)

load_env_files()


# fmt: off
class BaseConfig:
    DEBUG = False
    TESTING = False

    # Basic info
    APP_NAME = os.getenv('APP_NAME')
    APP_VERSION = os.getenv('APP_VERSION', '1.0.0')
    APP_ENV = os.getenv('APP_ENV', 'development')

    SECRET_KEY = os.getenv('SECRET_KEY', 'fool-said-i-you-do-not-know')

    # App settings
    POSTS_PER_PAGE = os.getenv('POSTS_PER_PAGE', 30)

    # Server settings
    HTTP_IP = os.getenv('HTTP_IP', '127.0.0.1')
    HTTP_PORT = int(os.getenv('HTTP_PORT', 8000))
    MAX_CONTENT_LENGTH = int(os.getenv('HTTP_MAX_BODY_SIZE', 10 * 1024 * 1024))

    # CORS settings
    CORS_ALLOWED_ORIGINS = os.getenv('CORS_ALLOWED_ORIGINS', '*')
    CORS_ALLOWED_METHODS = os.getenv('CORS_ALLOWED_METHODS', 'GET, POST, PUT, DELETE, OPTIONS')
    CORS_ALLOWED_HEADERS = os.getenv('CORS_ALLOWED_HEADERS', 'Content-Type, Authorization')
    # TODO: allow credentials only when specific origins are set
    CORS_ALLOW_CREDENTIALS = (os.getenv('CORS_ALLOW_CREDENTIALS', 'true').lower() == 'true')
    CORS_MAX_AGE = int(os.getenv('CORS_MAX_AGE', 86400))

    # Uploads settings
    UPLOAD_PATH = os.getenv('UPLOAD_PATH', os.path.join(PY_ROOT, 'uploads'))
    UPLOAD_URL = os.getenv('UPLOAD_URL', '/uploads')
    UPLOAD_IMAGE_FORMATS = ['png', 'jpg', 'jpeg', 'gif', 'webp']
    UPLOAD_THUMB_WIDTH = os.getenv('UPLOAD_THUMB_WIDTH', 200)

    # Database settings
    SQLALCHEMY_TRACK_MODIFICATIONS = False
    SQLALCHEMY_ECHO = False
    SQLALCHEMY_DATABASE_URI = 'sqlite:///app.db'
    AUTO_MIGRATE = (os.getenv('DATABASE_AUTO_MIGRATE', 'true').lower() == 'true')

    # Redis settings
    REDIS_URL = os.getenv('REDIS_URL', 'redis://localhost:6379/0')

    # Misc
    ABOUT_URL = os.getenv('ABOUT_URL')


class DevelopmentConfig(BaseConfig):
    DEBUG = True
    SQLALCHEMY_DATABASE_URI = os.getenv('DATABASE_URL', 'sqlite:///app-dev.db')


class ProductionConfig(BaseConfig):
    # sqlite
    SQLALCHEMY_DATABASE_URI = os.getenv('DATABASE_URL', 'sqlite:///app.db')


class TestConfig(BaseConfig):
    TESTING = True
    SQLALCHEMY_DATABASE_URI = 'sqlite:///:memory:'


config = ProductionConfig
env = os.getenv('FLASK_ENV')

if env == 'development':
    config = DevelopmentConfig
elif env == 'test':
    config = TestConfig
else:
    config = ProductionConfig
