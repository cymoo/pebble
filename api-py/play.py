from app.model import Post, db
from time import time
import os

os.environ['FLASK_ENV'] = 'development'
from wsgi import flask_app as app


def t1():
    with app.app_context():
        pass


if __name__ == '__main__':
    t1()
