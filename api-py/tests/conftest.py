import pytest
from app.config import TestConfig
from app import create_app
from app.model import db as _db


@pytest.fixture(scope="session")
def app():
    app = create_app(TestConfig)
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
