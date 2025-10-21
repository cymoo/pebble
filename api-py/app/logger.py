import logging
import os
from logging.config import dictConfig
from flask import has_request_context, request
from .util import get_real_and_safe_ip


class RequestFormatter(logging.Formatter):
    def format(self, record):
        record.url = 'N/A'
        record.remote_addr = 'N/A'
        record.method = 'N/A'

        if has_request_context():
            record.url = request.base_url or 'N/A'
            record.remote_addr = get_real_and_safe_ip() or 'N/A'
            record.method = request.method or 'N/A'

            # Optional: add additional request context information
            # record.endpoint = request.endpoint or 'N/A'
            # record.user_agent = request.user_agent.string or 'N/A'

        return super().format(record)


def setup_logger(config):
    # ensures that the specified log directory exists
    os.makedirs(os.path.dirname(config.LOG_FILE), exist_ok=True)

    logging_config = {
        'version': 1,
        'disable_existing_loggers': False,
        'formatters': {
            'simple': {
                'format': '%(asctime)s - %(levelname)s: %(message)s',
                'datefmt': '%Y-%m-%d %H:%M:%S',
            },
            'verbose': {
                'class': 'app.logger.RequestFormatter',
                'format': (
                    '%(asctime)s - %(levelname)s in %(name)s - '
                    '[%(remote_addr)s starting %(method)s %(url)s]: %(message)s'
                ),
                'datefmt': '%Y-%m-%d %H:%M:%S',
            },
        },
        'handlers': {
            'console': {
                'class': 'logging.StreamHandler',
                'formatter': 'simple',
                'stream': 'ext://sys.stdout',
            },
            'file': {
                'class': 'logging.handlers.RotatingFileHandler',
                'formatter': 'verbose',
                'filename': config.LOG_FILE,
                'maxBytes': config.LOG_MAX_BYTES,
                'backupCount': config.LOG_MAX_BACKUPS,
                'encoding': 'utf8',
            },
        },
        'root': {
            'level': config.LOG_LEVEL,
            'handlers': [config.LOG_TYPE],
        },
    }

    logging.config.dictConfig(logging_config)

    # reduce verbosity in production where log files are used
    if config.LOG_TYPE == 'file':
        logging.getLogger('werkzeug').setLevel(logging.WARNING)
