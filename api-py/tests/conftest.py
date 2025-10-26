import os

import pytest
from app.config import Config
from app import create_app
from app.model import db as _db


@pytest.fixture(scope="session")
def app():
    config = Config.from_env()
    database_url = os.environ.get("DATABASE_URL_TEST") or "sqlite:///:memory:"
    config.SQLALCHEMY_DATABASE_URI = database_url

    app = create_app(config)
    return app


@pytest.fixture(scope='session')
def db(app):
    with app.app_context():
        _db.create_all()
        yield _db
        _db.drop_all()


@pytest.fixture(scope='function')
def session(db):
    db.session.begin_nested()

    yield db.session

    db.session.rollback()
    db.session.remove()
