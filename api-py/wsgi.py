from app import create_app
from app.config import config

flask_app = create_app(config)
app = flask_app.wsgi_app
