import inspect
import pydantic
from functools import wraps
from typing import Callable, Literal, get_type_hints, Optional
from flask import abort, request
from pydantic import BaseModel


def rate_limit(max_count: int, expires: int = 60) -> Callable:
    """Decorator to limit the number of requests to a view function.
    Args:
        max_count (int): Maximum number of requests allowed within the expiration period.
        expires (int): Expiration time in seconds for the rate limit window. Default is 60 seconds.
    Returns:
        Callable: A decorator that can be applied to Flask view functions.
    Note:
        It requires a Redis instance to store the request counts. It's not ideal to have this dependency in middleware,
        but it's a simple and effective way to implement rate limiting in a distributed environment.
    Usage:
        @app.route('/some-endpoint')
        @rate_limit(max_count=5, expires=60)
        def some_view_function():
            ...
    """
    from .extension import rd

    def wrapper(view_func):
        @wraps(view_func)
        def inner(*args, **kw):
            key = 'rate:' + view_func.__qualname__
            pipe = rd.pipeline()
            pipe.set(key, 0, ex=expires, nx=True).incr(key)
            _, rv = pipe.execute()
            if rv > max_count:
                abort(429)
            else:
                return view_func(*args, **kw)

        return inner

    return wrapper


def validate(
    view_func: Optional[Callable] = None,
    *,
    type: Literal['json', 'query', 'form'] = 'json',
):
    """Decorator to validate request data using Pydantic models.
    The first parameter of the decorated view function must be annotated
    with a Pydantic model class. The decorator will parse and validate
    the incoming request data (JSON body, query parameters, or form data)
    against this model before passing it to the view function.
    Args:
        view_func: The view function to be decorated.
        type: The type of request data to validate.
            Can be 'json' for JSON body, 'query' for query parameters, or 'form' for form data.
            Default is 'json'.
    Returns:
        Callable: The decorated view function with request data validation.
    Raises:
        TypeError: If the first parameter is not annotated with a Pydantic model.
        ValidationError: If the request data fails validation.
    Usage:
        @app.post('/some-endpoint')
        @validate
        def some_view_function(payload: SomePydanticModel):
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
                raise ValidationError(_format_validation_error(err))

        return wrapper

    if view_func is None:
        return decorator
    return decorator(view_func)


def _format_validation_error(error: pydantic.ValidationError) -> str:
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
