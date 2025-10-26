import ipaddress
import os
from pathlib import Path
from typing import List, Optional
from dataclasses import dataclass


def get_env(key: str, default: str = "") -> str:
    return os.environ.get(key, default)


def get_env_int(key: str, default: int) -> int:
    value = os.environ.get(key)
    return int(value) if value else default


def get_env_bool(key: str, default: bool) -> bool:
    value = os.environ.get(key)
    if value is None:
        return default
    return value.lower() in ("true", "1", "yes", "on")


def get_env_list(key: str, default: Optional[List[str]] = None) -> List[str]:
    value = os.environ.get(key)
    if not value:
        return default or []
    return [item.strip() for item in value.split(",") if item.strip()]


@dataclass
class Config:
    """Application configuration that works seamlessly with Flask."""

    # Basic app info
    APP_NAME: str
    APP_VERSION: str
    APP_ENV: str

    # Flask core
    DEBUG: bool
    TESTING: bool
    SECRET_KEY: str

    # Application settings
    POSTS_PER_PAGE: int
    STATIC_URL: str
    STATIC_PATH: str

    # Server settings
    MAX_CONTENT_LENGTH: int

    # Database (Flask-SQLAlchemy compatible)
    SQLALCHEMY_DATABASE_URI: str
    SQLALCHEMY_TRACK_MODIFICATIONS: bool = False
    SQLALCHEMY_ENGINE_OPTIONS: dict = None

    # Redis
    REDIS_URL: str
    REDIS_PASSWORD: str
    REDIS_DB: int

    # Upload
    UPLOAD_FOLDER: str
    UPLOAD_URL: str
    ALLOWED_EXTENSIONS: set = None
    THUMB_WIDTH: int = 128

    # CORS
    CORS_ALLOWED_ORIGINS: List[str] = None
    CORS_ALLOWED_METHODS: List[str] = None
    CORS_ALLOWED_HEADERS: List[str] = None
    CORS_ALLOW_CREDENTIALS: bool = False
    CORS_MAX_AGE: int = 86400

    # Logging
    LOG_REQUESTS: bool = True

    def __post_init__(self):
        """Initialize computed fields and defaults."""
        if self.SQLALCHEMY_ENGINE_OPTIONS is None:
            pool_size = get_env_int("DATABASE_POOL_SIZE", 5)
            self.SQLALCHEMY_ENGINE_OPTIONS = {
                "pool_size": pool_size,
                "pool_pre_ping": True,
            }

        if self.ALLOWED_EXTENSIONS is None:
            formats = get_env_list(
                "UPLOAD_IMAGE_FORMATS", ["jpg", "jpeg", "png", "webp", "gif"]
            )
            self.ALLOWED_EXTENSIONS = set(formats)

        if self.CORS_ALLOWED_ORIGINS is None:
            self.CORS_ALLOWED_ORIGINS = []
        if self.CORS_ALLOWED_METHODS is None:
            self.CORS_ALLOWED_METHODS = []
        if self.CORS_ALLOWED_HEADERS is None:
            self.CORS_ALLOWED_HEADERS = []

    @classmethod
    def from_env(cls) -> "Config":
        """Load configuration from environment variables."""
        app_env = get_env("APP_ENV", "development")

        return cls(
            # Basic app info
            APP_NAME=get_env("APP_NAME", "Pebble"),
            APP_VERSION=get_env("APP_VERSION", "1.0.0"),
            APP_ENV=app_env,
            # Flask core
            DEBUG=(app_env == "development"),
            TESTING=(app_env == "testing"),
            SECRET_KEY=get_env("SECRET_KEY", "dev-secret-key-change-in-production"),
            # Application settings
            POSTS_PER_PAGE=get_env_int("POSTS_PER_PAGE", 30),
            STATIC_URL=get_env("STATIC_URL", "/static"),
            STATIC_PATH=get_env("STATIC_PATH", ""),
            # Server settings
            MAX_CONTENT_LENGTH=get_env_int("HTTP_MAX_BODY_SIZE", 1024 * 1024 * 10),
            # Database
            SQLALCHEMY_DATABASE_URI=get_env("DATABASE_URL", "sqlite:///app.db"),
            # Redis
            REDIS_URL=get_env("REDIS_URL", "localhost:6379"),
            REDIS_PASSWORD=get_env("REDIS_PASSWORD", ""),
            REDIS_DB=get_env_int("REDIS_DB", 0),
            # Upload
            UPLOAD_FOLDER=get_env("UPLOAD_PATH", "./uploads"),
            UPLOAD_URL=get_env("UPLOAD_URL", "/uploads"),
            THUMB_WIDTH=get_env_int("UPLOAD_THUMB_WIDTH", 128),
            # CORS
            CORS_ALLOWED_ORIGINS=get_env_list("CORS_ALLOWED_ORIGINS", []),
            CORS_ALLOWED_METHODS=get_env_list("CORS_ALLOWED_METHODS", []),
            CORS_ALLOWED_HEADERS=get_env_list("CORS_ALLOWED_HEADERS", []),
            CORS_ALLOW_CREDENTIALS=get_env_bool("CORS_ALLOW_CREDENTIALS", False),
            CORS_MAX_AGE=get_env_int("CORS_MAX_AGE", 86400),
            # Logging
            LOG_REQUESTS=get_env_bool("LOG_REQUESTS", True),
        )

    def validate(self) -> None:
        """Validate configuration and raise ValueError if invalid."""
        errors = []

        # Basic validation
        if not self.APP_NAME:
            errors.append("APP_NAME cannot be empty")
        if (
            not self.SECRET_KEY
            or self.SECRET_KEY == "dev-secret-key-change-in-production"
        ):
            if self.APP_ENV == "production":
                errors.append("SECRET_KEY must be set in production")

        # Application settings
        if self.POSTS_PER_PAGE <= 0 or self.POSTS_PER_PAGE > 1000:
            errors.append("POSTS_PER_PAGE must be between 1 and 1000")

        # Database
        if not self.SQLALCHEMY_DATABASE_URI:
            errors.append("SQLALCHEMY_DATABASE_URI cannot be empty")

        # Upload validation
        if not self.UPLOAD_FOLDER:
            errors.append("UPLOAD_FOLDER cannot be empty")
        else:
            path = Path(self.UPLOAD_FOLDER)
            try:
                path.mkdir(parents=True, exist_ok=True)
                test_file = path / ".write_test"
                test_file.write_text("test")
                test_file.unlink()
            except Exception as e:
                errors.append(
                    f"UPLOAD_FOLDER '{self.UPLOAD_FOLDER}' is not writable: {e}"
                )

        if self.THUMB_WIDTH <= 0 or self.THUMB_WIDTH > 4096:
            errors.append("THUMB_WIDTH must be between 1 and 4096")

        # Redis validation
        if not self.REDIS_URL:
            errors.append("REDIS_URL cannot be empty")
        if self.REDIS_DB < 0 or self.REDIS_DB > 15:
            errors.append("REDIS_DB must be between 0 and 15")

        if errors:
            raise ValueError(
                "Configuration validation failed:\n  - " + "\n  - ".join(errors)
            )
