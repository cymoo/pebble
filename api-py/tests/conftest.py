import os

import pytest
from app.config import TestConfig
from app.model import db as sa
from flask import Flask


@pytest.fixture
def app():
    app = Flask(__name__)
    app.config.from_object(TestConfig)
    sa.init_app(app)
    return app


@pytest.fixture
def db(app):
    with app.app_context():
        sa.create_all()
        yield sa
        sa.session.remove()
        sa.drop_all()
