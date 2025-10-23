import os
from flask import Flask
from flask_sqlalchemy import SQLAlchemy
from sqlalchemy import MetaData
from redis import Redis
from .lib.search import FullTextSearch


# https://stackoverflow.com/questions/45527323
naming_convention = {
    "ix": 'ix_%(column_0_label)s',
    "uq": "uq_%(table_name)s_%(column_0_name)s",
    "ck": "ck_%(table_name)s_%(column_0_name)s",
    "fk": "fk_%(table_name)s_%(column_0_name)s_%(referred_table_name)s",
    "pk": "pk_%(table_name)s",
}

db: SQLAlchemy = SQLAlchemy(metadata=MetaData(naming_convention=naming_convention))
rd: Redis = None  # type: ignore
fts: FullTextSearch = None  # type: ignore


def init_app(app: Flask) -> None:
    global db, rd, fts

    db.init_app(app)
    # Inject app into db for convenience, it's not a common practice though
    db.app = app

    rd = Redis.from_url(app.config['REDIS_URL'], decode_responses=True)
    fts = FullTextSearch(rd, 'fts:')


def run_migration(app: Flask, db: SQLAlchemy) -> None:
    """Run database migrations using Flask-Migrate.
    If migrations folder does not exist or migration fails, create all tables directly.
    """
    from flask_migrate import Migrate, upgrade
    from .logger import logger
    from concurrent.futures import ProcessPoolExecutor

    _ = Migrate(app, db)

    with app.app_context():
        if not os.path.exists('migrations'):
            logger.info("Migrations folder not found")
            logger.info("Run the following command to initialize migrations:")
            logger.info("    flask db init")
            logger.info("    flask db migrate -m 'Initial migration'")
            logger.info("    flask db upgrade")
            logger.info("Now creating the database tables directly.")
            db.create_all()
        else:
            try:
                # NOTE: flask_migrate.upgrade breaks logging!!!
                # So we run it in a separate process to avoid messing up the main process's logger.
                # A bitter lesson learned: reduce dependencies as much as possible.
                with ProcessPoolExecutor() as executor:
                    future = executor.submit(upgrade)
                    future.result(timeout=10)  # wait for max 10 seconds

                logger.info("Database migrated to latest version")
            except Exception as e:
                logger.error(f"Database migration failed: {e}")
                logger.info("Creating database tables directly.")
                db.create_all()
