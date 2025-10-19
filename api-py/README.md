# Flask API

The backend built with Python, Flask, SQLite, SQLAlchemy and Redis.

## Getting Started

To begin with this project:

### Create a Virtual Environment and Install Dependencies

```bash
python3 -m venv venv
source venv/bin/activate
pip install -r requirements.txt
```

### Run Redis

```bash
redis-server
```

### Starting the Application

- Run in development

```bash
PEBBLE_PASSWORD=xxx flask run
```

- Run in production

```bash
PEBBLE_PASSWORD=xxx gunicorn -k gevent -b :8000 wsgi:app
```

NOTE: The `PEBBLE_PASSWORD` variable is used for login. Ensure it is complex and securely stored in production.

### Database Migration

To create the sqlite database and tables if missing:

```bash
flask create_tables
```

1. Init migration

```bash
flask db init
```

2. Autogenerate a new revision file

```bash
flask db migrate
```

3. Upgrade to a new version

```bash
flask db upgrade
```

For more usage details about migration, run `flask db --help`

### Test

```bash
pytest .
```
