from app import create_app
from app.config import Config

cfg = Config.from_env()
flask_app = create_app(cfg)
app = flask_app.wsgi_app
