import logging
import sys
from logging.handlers import RotatingFileHandler
from pathlib import Path
from flask import has_request_context, request
from ipaddress import ip_address
from typing import Optional

NA = '-'


class RequestFormatter(logging.Formatter):
    """Custom formatter to include Flask request context info in logs"""

    def format(self, record):
        record.url = NA
        record.remote_addr = NA
        record.method = NA

        if has_request_context():
            record.url = request.base_url or NA

            record.remote_addr = get_real_ip() or NA
            record.method = request.method or NA

            # Optional: include more request info if needed
            # record.endpoint = request.endpoint or NA
            # record.user_agent = request.user_agent.string or NA

        return super().format(record)


# Global logger instance
logger = logging.getLogger('app')


def config_logger(
    level=logging.INFO,
    console_output=True,
    log_file='logs/app.log',
    include_request_context=False,
    max_bytes=10485760,  # 10MB
    max_backups=10,
):
    """
    Configure the global logger with specified settings.
    Arguments:
        level: Logging level (e.g., logging.INFO, logging.DEBUG)
        console_output: Whether to log to console (stdout)
        log_file: Path to log file. If None, file logging is disabled.
        include_request_context: Whether to include Flask request context in logs
        max_bytes: Maximum size of log file before rotation
        backup_count: Number of backup log files to keep
    Returns:
        Configured logger instance
    """
    # Clear existing handlers
    if logger.handlers:
        logger.handlers.clear()

    logger.setLevel(level)

    # Formatter based on whether to include request context
    if include_request_context:
        formatter = RequestFormatter(
            fmt='[%(asctime)s] %(levelname)s %(remote_addr)s %(method)s %(url)s - %(module)s.%(funcName)s:%(lineno)d - %(message)s',
            datefmt='%Y-%m-%d %H:%M:%S',
        )
    else:
        formatter = logging.Formatter(
            fmt='[%(asctime)s] %(levelname)s in %(module)s.%(funcName)s:%(lineno)d - %(message)s',
            datefmt='%Y-%m-%d %H:%M:%S',
        )

    # Console handler
    if console_output:
        console_handler = logging.StreamHandler(sys.stdout)
        console_handler.setLevel(level)
        console_handler.setFormatter(formatter)
        logger.addHandler(console_handler)

    # File handler - use RotatingFileHandler for log rotation
    if log_file:
        # Ensure log directory exists
        log_path = Path(log_file)
        log_path.parent.mkdir(parents=True, exist_ok=True)

        file_handler = RotatingFileHandler(
            log_file, maxBytes=max_bytes, backupCount=max_backups, encoding='utf-8'
        )
        file_handler.setLevel(level)
        file_handler.setFormatter(formatter)
        logger.addHandler(file_handler)

    # Prevent log messages from being propagated to the root logger
    logger.propagate = False

    return logger


def get_real_ip() -> Optional[str]:
    """Securely obtain the client IP address.

    Supports IPv4/IPv6 and handles multiple proxy layers.
    Priority:
    1. X-Forwarded-For
    2. X-Real-IP
    3. remote_addr
    """

    ip_candidates = []

    # X-Forwarded-For may contain multiple IPs
    forwarded_for = request.headers.get('X-Forwarded-For', '').split(',')
    ip_candidates.extend([ip.strip() for ip in forwarded_for if ip.strip()])

    # add X-Real-IP
    real_ip = request.headers.get('X-Real-IP')
    if real_ip:
        ip_candidates.append(real_ip)

    # add remote_addr
    if request.remote_addr:
        ip_candidates.append(request.remote_addr)

    # validate and return the first valid IP
    for candidate in ip_candidates:
        try:
            # validate IP
            ip = ip_address(candidate)
            # exclude private/reserved addresses
            if not (ip.is_private or ip.is_reserved):
                return str(ip)
        except ValueError:
            continue

    return None
