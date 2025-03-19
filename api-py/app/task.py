from datetime import datetime, timedelta, UTC
from app.model import db, Post

from huey import crontab
from huey.contrib.mini import MiniHuey, logger

huey = MiniHuey()


@huey.task(crontab(minute='0', hour='3'))
def clear_posts():
    with db.app.app_context():
        sixty_days_ago = datetime.now(UTC) - timedelta(days=60)
        deleted_count = Post.query.filter(
            Post.deleted_at < int(sixty_days_ago.timestamp() * 1000)
        ).delete()
        db.session.commit()

        if deleted_count:
            logger.info(f'[Daily] Deleted {deleted_count} posts.')
