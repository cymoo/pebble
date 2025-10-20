import os
from pathlib import PurePath

from app.util import load_env_files

BASE_DIR = str(PurePath(__file__).parent.parent.parent)

load_env_files()


class BaseConfig:
    STATIC_FOLDER = os.path.join(BASE_DIR, 'build')
    UPLOAD_FOLDER = os.path.join(BASE_DIR, 'data/uploads')

    MAX_CONTENT_LENGTH = 10 * 1024 * 1024
    IMAGE_FORMATS = ('.png', '.jpg', '.jpeg', '.gif', '.webp')
    THUMBNAIL_WIDTH = 128

    POST_NUM_PER_PAGE = 30

    # CORS
    ACCESS_CONTROL_ALLOW_ORIGIN = '*'
    ACCESS_CONTROL_ALLOW_METHODS = '*'
    ACCESS_CONTROL_ALLOW_HEADERS = '*'

    # Logging
    LOG_TYPE = 'console'
    LOG_LEVEL = "DEBUG"
    LOG_FILE_PATH = 'logs/app.log'
    LOG_MAX_BYTES = 10 * 1024 * 1024  # 10MB
    LOG_BACKUP_COUNT = 10

    # Redis
    REDIS = {
        'host': 'localhost',
        'port': 6379,
        'decode_responses': True,
    }

    # Misc
    ABOUT_URL = os.environ.get('ABOUT_URL')


class DevelopmentConfig(BaseConfig):
    # display query for debug
    # SQLALCHEMY_ECHO = True
    SQLALCHEMY_TRACK_MODIFICATIONS = False
    SQLALCHEMY_DATABASE_URI = f'sqlite:///{os.path.join(BASE_DIR, "data/app-dev.db")}'


class TestConfig(BaseConfig):
    SQLALCHEMY_DATABASE_URI = 'sqlite:///:memory:'

    REDIS = {
        'host': 'localhost',
        'port': 6379,
        'db': 15,
        'decode_responses': True,
    }


class ProductionConfig(BaseConfig):
    # sqlite
    SQLALCHEMY_TRACK_MODIFICATIONS = False
    SQLALCHEMY_DATABASE_URI = f'sqlite:///{os.path.join(BASE_DIR, "data/app.db")}'

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
