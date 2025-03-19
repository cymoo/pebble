import inspect
import random
import re
import string
import warnings
from datetime import timedelta, datetime, timezone
from functools import partial, wraps
from ipaddress import ip_address
from types import NoneType
from typing import Callable, TypeAlias, Literal, get_type_hints, Optional
from uuid import uuid4

import pydantic
from PIL import Image
from flask import abort, request, current_app as app
from pydantic import BaseModel


def deprecated(func):
    def wrapper(*args, **kwargs):
        warnings.warn(
            f"{func.__name__} is deprecated",
            category=DeprecationWarning,
            stacklevel=2,
        )
        return func(*args, **kwargs)

    return wrapper


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


def ms_now() -> int:
    return int(datetime.now(timezone.utc).timestamp() * 1000)


def str_to_datetime(datestr: str, offset: int, end_of_day=False) -> datetime:
    """
    Convert a date string to a `datetime` object with timezone information

    Args:
        datestr: The date string in "yyyy-MM-dd" format
        offset: Timezone offset in minutes
        end_of_day: Whether to use the end time of the day (defaults to false, which means the start time of the day)
    """
    if abs(offset) > 1440:
        raise ValueError(
            f'Timezone offset must be between -1440 and 1440 minutes: {offset}'
        )

    # Parse the date string
    local_datetime = datetime.strptime(datestr, "%Y-%m-%d")

    # Create time component
    local_datetime = local_datetime.replace(
        hour=23 if end_of_day else 0,
        minute=59 if end_of_day else 0,
        second=59 if end_of_day else 0,
        microsecond=999000 if end_of_day else 0,
    )

    # Create timezone offset
    offset_hours, offset_minutes = divmod(offset, 60)
    custom_timezone = timezone(timedelta(hours=offset_hours, minutes=offset_minutes))

    # Add timezone info for `datetime` object
    return local_datetime.replace(tzinfo=custom_timezone)


class classproperty:  # noqa
    """A decorator to define a class-level property.

    This custom descriptor allows you to create properties that are accessed
    directly on the class (rather than instances) while still being able to
    use a method to compute the property value dynamically.
    """

    def __init__(self, method) -> None:
        self.method = method

    def __get__(self, instance, owner):
        return self.method(owner)


def random_string(length: int = 10, str_type: str = 'all') -> str:
    """Generates a random string of the specified length and type.

    It creates a string with random characters based on the specified type,
    which can include digits, letters, uppercase, lowercase, or a combination of these.
    """

    choices = {
        'digit': string.digits,
        'letter': string.ascii_letters,
        'uppercase': string.ascii_uppercase,
        'lowercase': string.ascii_lowercase,
        'upper_digit': string.digits + string.ascii_uppercase,
        'lower_digit': string.digits + string.ascii_lowercase,
        'all': string.digits + string.ascii_letters,
    }
    return ''.join(random.choice(choices[str_type]) for _ in range(length))


random_digits = partial(random_string, str_type='digit')
random_letters = partial(random_string, str_type='letter')
random_upper_letters = partial(random_string, str_type='uppercase')
random_lower_letters = partial(random_string, str_type='lowercase')


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
    >>> add_suffix("foo.jpg", suffix=".thumb")
    'foo.thumb.jpg'
    """

    parts = filename.rsplit('.', maxsplit=1)
    if len(parts) == 1:
        basename, ext = parts[0], ''
    else:
        basename, ext = parts[0], '.' + parts[1]
    return f'{basename}{suffix}{ext}'


def get_parent_path(path: str) -> str:
    """
    >>> get_parent_path('a')
    ''
    >>> get_parent_path('a/b')
    'a'
    >>> get_parent_path('a/b/c')
    'a/b'
    """

    if '/' not in path:
        return ''
    else:
        return path.rsplit('/', maxsplit=1)[0]


def get_real_ip() -> Optional[str]:
    """A simple method to obtain the real IP address:
    Priority:
    1. X-Forwarded-For
    2. X-Real-IP
    3. remote_addr
    """

    # try to get from X-Forwarded-For
    forwarded_for = request.headers.get('X-Forwarded-For')
    if forwarded_for:
        return forwarded_for.split(',')[0].strip()

    # try to get from X-Real-IP
    real_ip = request.headers.get('X-Real-IP')
    if real_ip:
        return real_ip

    # fallback to remote_addr
    return request.remote_addr


def get_real_and_safe_ip() -> Optional[str]:
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


def limit_request(count: int, interval: int = 60) -> Callable:
    """Limits the number of times a function can be called within a specified time interval.

    This decorator can be used to restrict the number of requests to a view function,
    ensuring that it is not called more than `count` times within the specified `interval` (in seconds).
    If the function exceeds the limit, an HTTP 429 (Too Many Requests) error will be returned.
    """

    def wrapper(view_func):
        @wraps(view_func)
        def inner(*args, **kw):
            key = 'rate:' + view_func.__qualname__
            pipe = app.rd.pipeline()  # noqa
            pipe.set(key, 0, ex=interval, nx=True).incr(key)
            _, rv = pipe.execute()
            if rv > count:
                abort(429)
            else:
                return view_func(*args, **kw)

        return inner

    return wrapper


BasicTypes: TypeAlias = int | float | str | bool | NoneType


# fmt: off
def validate(view_func: Optional[Callable]=None, *, type: Literal['json', 'query', 'form'] = 'json'):  # noqa
    """Decorator to validate request data using Pydantic.

     Examples:
        @app.post('/create-user')
        @validate
        def create_user(payload: UserCreate):
            ...

        @app.post('/update-user/<int:id>')
        @validate(type='form')
        def update_user(payload: UserUpdate, id: int):
            ...
    """

    from .exception import ValidationError

    def decorator(func):
        hints = get_type_hints(func)

        params = inspect.signature(func).parameters
        first_param_name = next(iter(params), None)

        if first_param_name is None:
            raise TypeError('view function must have at least one parameter')

        model_cls = hints.get(first_param_name)

        if not model_cls or not issubclass(model_cls, BaseModel):
            raise TypeError(
                f"first parameter '{first_param_name}' must be annotated with a Pydantic model"
            )

        if type not in ('json', 'query', 'form'):
            raise ValueError('request type must be json, query, or form')

        @wraps(func)
        def wrapper(*args, **kwargs):
            try:
                if type == 'query':
                    validated_data = model_cls(**request.args)
                elif type == 'form':
                    validated_data = model_cls(**request.form)
                else:
                    validated_data = model_cls(**request.json)

                return func(validated_data, *args, **kwargs)
            except pydantic.ValidationError as err:
                raise ValidationError(format_validation_error(err))

        return wrapper

    if view_func is None:
        return decorator
    return decorator(view_func)


def format_validation_error(error: pydantic.ValidationError) -> str:
    """Formats a Pydantic validation error and generates a human-readable
    string that summarizes the validation issues for each field, including nested fields.
    """

    field_errors: dict[str, list[str]] = {}

    for err in error.errors():
        # Handle errors for nested fields
        location = " -> ".join(str(loc) for loc in err["loc"])
        message = err["msg"]

        if location not in field_errors:
            field_errors[location] = []
        field_errors[location].append(message)

    parts = []

    for field, messages in field_errors.items():
        error_text = f"{field}: {' '.join(f'[{msg}]' for msg in messages)}"
        parts.append(error_text)

    return "; ".join(parts)


def highlight_html(tokens: list[str], html: str) -> str:
    """
    Mark all occurrences of tokens in HTML text with <mark> tags,
    avoiding replacements in HTML tags and their attributes

    Args:
        tokens: List of tokens to be marked
        html: Original HTML text

    Returns:
        HTML text with tokens marked only in text content
    """
    if not tokens:
        return html

    # Add word boundaries for English tokens
    patterns = []
    for token in sorted(tokens, key=len, reverse=True):
        if any('\u4e00' <= char <= '\u9fff' for char in token):  # Chinese
            patterns.append(re.escape(token))
        else:  # English
            patterns.append(fr'\b{re.escape(token)}\b')

    # Single regex to match tokens but ignore HTML tags
    pattern = f'(?:<[^>]*>)|({"|".join(patterns)})'
    return re.sub(pattern, lambda m: m.group(0) if m.group(1) is None else f'<mark>{m.group(1)}</mark>', html)


if __name__ == '__main__':
    import doctest

    doctest.testmod()
