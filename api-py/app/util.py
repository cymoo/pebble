import os
import re
import warnings
from datetime import timedelta, datetime, timezone
from typing import Literal
from uuid import uuid4
from PIL import Image
from dotenv import load_dotenv


def deprecated(func):
    """A decorator to mark functions as deprecated.
    It will issue a warning when the function is called.
    """

    def wrapper(*args, **kwargs):
        warnings.warn(
            f"{func.__name__} is deprecated",
            category=DeprecationWarning,
            stacklevel=2,
        )
        return func(*args, **kwargs)

    return wrapper


def load_env_files(env: Literal['dev', 'prod', 'test'] | None = None) -> None:
    """
    Load environment files based on env type with the following priority:
    1. .env (base configuration)
    2. .env.[environment] (environment-specific configuration)
    3. .env.local (local overrides)

    Args:
        env (str, optional): Explicitly specify the environment (e.g., 'dev', 'prod', 'test').
            If not provided, uses the FLASK_ENV environment variable (defaults to 'development').
    """
    # 1. Determine environment (priority: argument > FLASK_ENV variable > default 'development')
    env = (env or os.getenv("FLASK_ENV", "development")).lower()

    # 2. Load base .env file if exists
    if os.path.exists(".env"):
        load_dotenv(".env")

    # 3. Load environment-specific file
    env_mappings = {
        "dev": ".env.dev",
        "development": ".env.dev",
        "prod": ".env.prod",
        "production": ".env.prod",
        "test": ".env.test",
    }

    env_file = env_mappings.get(env)
    if env_file and os.path.exists(env_file):
        load_dotenv(env_file, override=True)

    # 4. Always load .env.local last for final overrides
    if os.path.exists(".env.local"):
        load_dotenv(".env.local", override=True)


class Missing:
    """A Singleton which indicates a value does not exist.

    >>> Missing() == Missing()
    True
    """

    _instance = None

    def __new__(cls, *args, **kw):
        if cls._instance is None:
            cls._instance = super().__new__(cls, *args, **kw)
        return cls._instance

    def __str__(self):
        return '<Missing>'

    __repr__ = __str__


missing = Missing()


def utc_now_ms() -> int:
    return int(datetime.now(timezone.utc).timestamp() * 1000)


def parse_date_with_timezone(date_str: str, utc_offset: int, at_end=False) -> datetime:
    """Parses a date string with a specified timezone offset.
    Args:
        date_str (str): Date string in "YYYY-MM-DD" format.
        utc_offset (int): Timezone offset in minutes from UTC.
        at_end (bool): If True, set time to 23:59:59.999, else to 00:00:00.000.
    Returns:
        datetime: A timezone-aware datetime object.
    Raises:
        ValueError: If the date format is invalid or timezone offset is out of range.
    """

    if abs(utc_offset) > 1440:
        raise ValueError(
            f'Timezone offset must be between -1440 and 1440 minutes: {utc_offset}'
        )

    # Parse the date string
    local_datetime = datetime.strptime(date_str, "%Y-%m-%d")

    # Create time component
    local_datetime = local_datetime.replace(
        hour=23 if at_end else 0,
        minute=59 if at_end else 0,
        second=59 if at_end else 0,
        microsecond=999000 if at_end else 0,
    )

    # Create timezone offset
    offset_hours, offset_minutes = divmod(utc_offset, 60)
    custom_timezone = timezone(timedelta(hours=offset_hours, minutes=offset_minutes))

    # Add timezone info for `datetime` object
    return local_datetime.replace(tzinfo=custom_timezone)


def gen_thumbnail(img: Image.Image, width: int) -> Image.Image:
    """Generates a thumbnail image with the specified width while maintaining the aspect ratio."""

    w_percent = width / img.width
    height = int(img.height * w_percent)
    return img.resize((width, height), Image.Resampling.LANCZOS)  # noqa


# not 汉字, digit, alphabet, -, _, .
INVALID_CHAR_PATTERN = re.compile(r'[^\w\-.\u4e00-\u9fa5]+')


def gen_secure_filename(filename: str, uuid_length=8) -> str:
    """Generates a secure, sanitized filename by replacing illegal characters with underscores
    and appending a unique suffix to avoid naming conflicts.

    It ensures the filename contains only valid characters (Chinese characters,
    digits, alphabets, hyphens, underscores, and periods) and appends an uuid to make it secure for storage.

    >>> gen_secure_filename("foo$&.jpg").startswith("foo_")
    True
    >>> '$&' not in gen_secure_filename("foo$&.jpg")
    True
    >>> gen_secure_filename("中文.jpg").startswith("中文")
    True
    """

    if not filename:
        raise ValueError('filename cannot be empty')

    if 8 > uuid_length > 32:
        raise ValueError('uuid_length must be between 8 and 32')

    filename = INVALID_CHAR_PATTERN.sub('_', filename)
    return add_suffix(filename, suffix='.' + uuid4().hex[0:uuid_length])


def add_suffix(filename: str, suffix: str) -> str:
    """
    Add a suffix to a filename before the file extension.
    >>> add_suffix("foo.jpg", suffix=".thumb")
    'foo.thumb.jpg'
    """

    parts = filename.rsplit('.', maxsplit=1)
    if len(parts) == 1:
        basename, ext = parts[0], ''
    else:
        basename, ext = parts[0], '.' + parts[1]
    return f'{basename}{suffix}{ext}'


def replace_prefix(s: str, from_str: str, to: str) -> str:
    """Replace the prefix of a string if it matches a given substring.
    >>> replace_prefix('foobar', 'foo', 'baz')
    'bazbar'
    >>> replace_prefix('foobar', 'bar', 'baz')
    'foobar'
    """
    if s.startswith(from_str):
        return to + s[len(from_str) :]
    else:
        return s


if __name__ == '__main__':
    import doctest

    doctest.testmod()
