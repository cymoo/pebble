import re
from datetime import datetime, timedelta
from typing import Optional, Tuple, Self, Iterable, Literal

from flask import abort
from flask_sqlalchemy import SQLAlchemy
from sqlalchemy.dialects.sqlite import insert

from sqlalchemy import MetaData, func, or_, text
from sqlalchemy.orm import backref, subqueryload, Query

from .util import deprecated, ms_now, replace_from_start

# https://stackoverflow.com/questions/45527323
naming_convention = {
    "ix": 'ix_%(column_0_label)s',
    "uq": "uq_%(table_name)s_%(column_0_name)s",
    "ck": "ck_%(table_name)s_%(column_0_name)s",
    "fk": "fk_%(table_name)s_%(column_0_name)s_%(referred_table_name)s",
    "pk": "pk_%(table_name)s",
}


db = SQLAlchemy(metadata=MetaData(naming_convention=naming_convention))

HASH_PATTERN = re.compile(r'<span class="hash-tag">#(.+?)</span>')

# https://stackoverflow.com/questions/25668092
tag_post_assoc = db.Table(
    'tag_post_assoc',
    db.Column(
        'tag_id',
        db.Integer,
        db.ForeignKey('tags.id', ondelete='CASCADE'),
        primary_key=True,
    ),
    db.Column(
        'post_id',
        db.Integer,
        db.ForeignKey('posts.id', ondelete='CASCADE'),
        primary_key=True,
    ),
)


class Tag(db.Model):
    __tablename__ = 'tags'
    id = db.Column(db.Integer, primary_key=True, nullable=False)
    name = db.Column(db.String(32), nullable=False, unique=True)
    sticky = db.Column(db.Boolean, nullable=False, default=False)

    created_at = db.Column(db.Integer, nullable=False, default=ms_now)
    updated_at = db.Column(
        db.Integer,
        nullable=False,
        default=ms_now,
        onupdate=ms_now,
    )

    posts = db.relationship(
        'Post',
        secondary=tag_post_assoc,
        back_populates='tags',
        lazy='dynamic',
        cascade='all, delete',
    )

    @classmethod
    def find_or_create(cls, name: str) -> Self:
        tag = Tag.find_by_name(name)
        if not tag:
            tag = Tag(name=name)
            tag.save()
        return tag

    @classmethod
    def find_by_name(cls, name: str) -> Optional[Self]:
        return Tag.query.filter_by(name=name).first()

    @staticmethod
    def get_all_with_post_count() -> list[Tuple[str, bool, int]]:
        """
        Retrieves all tags with their sticky status and associated post count.

        Returns:
            List of tuples containing (name: str, sticky: bool, post_count: int).
            Post count includes posts associated with the tag or its sub-tags.
        """
        result = db.session.execute(
            text(
                """
                WITH tag_posts AS (
                    SELECT t.name AS tag_name, p.id AS post_id
                    FROM tags t
                    JOIN tag_post_assoc tpa ON t.id = tpa.tag_id
                    JOIN posts p ON tpa.post_id = p.id
                    WHERE p.deleted_at IS NULL
                )
                SELECT t.name AS name,
                       t.sticky AS sticky,
                       COUNT(DISTINCT tp.post_id) AS post_count
                FROM tags t
                LEFT JOIN tag_posts tp ON tp.tag_name = t.name OR tp.tag_name LIKE (t.name || '/%')
                GROUP BY t.name
            """
            )
        )

        return [(row.name, row.sticky, row.post_count) for row in result]

    @classmethod
    def count(cls) -> int:
        # NOTE: Avoid using `Tag.query.count()`, the generated SQL is so good...
        return db.session.query(func.count(Tag.id)).scalar()

    @classmethod
    def insert_or_update(cls, name: str, sticky: bool) -> None:
        now = ms_now()
        stmt = (
            insert(Tag)
            .values(name=name, sticky=sticky, created_at=now, updated_at=now)
            .on_conflict_do_update(
                index_elements=['name'], set_={'sticky': sticky, 'updated_at': now}
            )
        )
        db.session.execute(stmt)
        db.session.commit()

    @classmethod
    def rename_or_merge(cls, name: str, new_name: str) -> None:
        """Rename a tag, and if the new tag already exists, merge the tags."""

        if name == new_name:
            return

        if new_name.startswith(name) and new_name.count('/') > name.count('/'):
            abort(400, f'cannot move "{name}" to a subtag of itself "{new_name}"')

        source_tag = Tag.find_or_create(name=name)
        target_tag = Tag.find_by_name(new_name)

        for descendant in source_tag.descendants:
            new_descendant_name = replace_from_start(descendant.name, name, new_name)
            new_descendant = Tag.find_by_name(new_descendant_name)
            if new_descendant:
                descendant._merge(new_descendant)
            else:
                descendant._rename(new_descendant_name)

        if target_tag:
            source_tag._merge(target_tag)
        else:
            source_tag._rename(new_name)

        db.session.commit()

    def _rename(self, new_name: str) -> None:
        old_name = self.name
        self.name = new_name
        db.session.add(self)

        for post in self.posts.all():
            post.content = post.content.replace(f'>#{old_name}<', f'>#{new_name}<')
            db.session.add(post)

    def _merge(self, new_tag: Self) -> None:
        old_name = self.name
        new_name = new_tag.name

        for post in self.posts.all():
            post.tags = [tag for tag in post.tags if tag.id != self.id] + [new_tag]
            post.content = post.content.replace(f'>#{old_name}<', f'>#{new_name}<')
            db.session.add(post)

    @property
    def descendants(self) -> list[Self]:
        """Retrieve all descendants of the tag."""
        return Tag.query.filter(Tag.name.startswith(self.name + '/')).all()

    @property
    def post_count(self) -> int:
        """Count the number of direct posts under the tag."""
        return (
            self.posts.filter(Post.deleted_at.is_(None))
            .with_entities(func.count(Post.id))
            .scalar()
        )

    def delete(self) -> None:
        """Soft delete all posts under this tag and its descendant tags."""

        deleted_at = ms_now()
        for tag in [self] + self.descendants:
            for post in tag.posts:  # noqa
                post.deleted_at = deleted_at
                db.session.add(post)

        db.session.commit()

    def restore(self) -> None:
        """Restore all posts under this tag and its descendant tags."""

        for tag in [self] + self.descendants:
            for post in tag.posts:  # noqa
                post.deleted_at = None
                db.session.add(post)
        db.session.commit()

    def save(self) -> None:
        db.session.add(self)
        db.session.commit()

    def __repr__(self) -> str:
        return f'<Tag id={self.id} name="{self.name}">'


class Post(db.Model):
    __tablename__ = 'posts'

    id = db.Column(db.Integer, primary_key=True, nullable=False)

    content = db.Column(db.Text, nullable=False)
    files = db.Column(db.Text)
    color = db.Column(db.String(8), index=True)  # red, green, blue

    shared = db.Column(db.Boolean, nullable=False, default=False)

    deleted_at = db.Column(db.Integer)
    created_at = db.Column(db.Integer, nullable=False, default=ms_now)
    updated_at = db.Column(
        db.Integer,
        nullable=False,
        default=ms_now,
        onupdate=ms_now,
    )

    # https://stackoverflow.com/questions/51335298/concepts-of-backref-and-back-populate-in-sqlalchemy
    tags = db.relationship('Tag', secondary=tag_post_assoc, back_populates='posts')

    parent_id = db.Column(db.Integer, db.ForeignKey('posts.id', ondelete='SET NULL'))
    parent = db.relationship(
        'Post',
        backref=backref('children', lazy='dynamic'),
        remote_side=[id],
    )
    # NOTE: Add an extra field to simplify queries and improve speed.
    children_count = db.Column(db.Integer, nullable=False, default=0)

    # `__init__` will only be called when creating a new record,
    # it will not be called when constructing an instance from existing records in the database.
    # https://stackoverflow.com/questions/16156650/sqlalchemy-init-not-running
    def __init__(self, **kwargs):
        content, parent_id = kwargs.get('content'), kwargs.get('parent_id')

        tags = [Tag.find_or_create(tag) for tag in HASH_PATTERN.findall(content)]
        super().__init__(**kwargs, tags=tags)

        if parent_id:
            parent = db.get_or_404(Post, parent_id, description='parent not exist')
            parent.children_count += 1
            db.session.add(parent)

    @classmethod
    def get_daily_counts(cls, start_date: datetime, end_date: datetime) -> list[int]:
        """Get the number of posts for each day within a specified date range."""

        # Convert `datetime` to timestamp (in milliseconds)
        start_timestamp = int(start_date.timestamp() * 1000)
        end_timestamp = int(end_date.timestamp() * 1000)

        offset = int(start_date.utcoffset().total_seconds())

        # Use SQLite functions to group and count posts by date
        daily_counts = (
            db.session.query(
                func.date(Post.created_at / 1000 + offset, 'unixepoch').label('date'),
                func.count(Post.id).label('count'),
            )
            .filter(Post.deleted_at.is_(None))
            .filter(Post.created_at.between(start_timestamp, end_timestamp))
            .group_by('date')
            .all()
        )

        daily_counts_map = dict(daily_counts)

        # Get daily post counts
        days = (end_date.date() - start_date.date()).days + 1
        return [
            daily_counts_map.get(
                (start_date.date() + timedelta(days=i)).strftime('%Y-%m-%d'), 0
            )
            for i in range(days)
        ]

    @classmethod
    def get_active_days(cls) -> int:  # noqa
        """Count the total number of active days."""

        return (
            Post.query.filter(Post.deleted_at.is_(None))
            .with_entities(
                func.count(
                    func.distinct(func.date(Post.created_at / 1000, 'unixepoch'))
                )
            )
            .scalar()
        )

    @classmethod
    def find_by_ids(cls, ids: Iterable[int]) -> list[Self]:
        query = (
            db.session.query(Post)
            # `subqueryload` generates a subquery for the related data,
            # loads all the required related data into memory in a single query,
            # and then SQLAlchemy associates this data with the main query result set, avoiding "N+1" query.
            .options(subqueryload(Post.tags))  # noqa
            .filter(Post.deleted_at.is_(None))
            .filter(Post.id.in_(ids))
        )

        return query.all()

    @classmethod
    def filter_posts(
        cls,
        *,
        cursor: Optional[float] = None,
        deleted: bool = False,
        parent_id: Optional[int] = None,
        color: Optional[str] = None,  # noqa
        tag: Optional[str] = None,
        start_date: int | None = None,
        end_date: int | None = None,
        shared: Optional[bool] = None,
        has_files: bool | None = None,
        order_by: Literal['created_at', 'updated_at', 'deleted_at'] = 'created_at',
        ascending: bool = False,
        per_page: int = 20,
    ) -> list[Self]:
        """Find posts based on various optional filters and pagination parameters."""

        # `subqueryload`: load the relationship using a separate query.
        # The default is `lazyload`, meaning a query is triggered each time `post.tags` is accessed.
        # Other common modes include `joinedload` and `contains_eager`.
        # https://python.plainenglish.io/relationships-with-sqlalchemy-958b7358e16
        query = db.session.query(Post).options(subqueryload(Post.tags))  # noqa

        if deleted:
            query = query.filter(Post.deleted_at.isnot(None))
        else:
            query = query.filter(Post.deleted_at.is_(None))

        if parent_id is not None:
            query = query.filter(Post.parent_id == parent_id)

        if color:
            query = query.filter(Post.color == color)

        if tag:
            # NOTE: When using limit with join, the result may be fewer than the limit.
            # It can be resolved by using a subquery.
            # https://dncrews.com/limit-and-offset-can-work-with-join-f03327fa2ad3
            query = query.join(Post.tags).filter(
                or_(Tag.name == tag, Tag.name.startswith(tag + '/'))
            )

        if start_date:
            query = query.filter(Post.created_at >= start_date)

        if end_date:
            query = query.filter(Post.created_at <= end_date)

        if shared is not None:
            if shared:
                query = query.filter(Post.shared)
            else:
                query = query.filter(~Post.shared)

        if has_files is not None:
            if has_files:
                query = query.filter(Post.files.isnot(None))
            else:
                query = query.filter(Post.files.is_(None))

        order_field = getattr(Post, order_by)
        query = query.order_by(order_field.asc() if ascending else order_field.desc())

        if cursor:
            query = query.filter(
                order_field > cursor if ascending else order_field < cursor
            )

        query = query.limit(per_page)

        return query.all()

    @classmethod
    def clear_all(cls) -> list[int]:
        posts_to_delete = (
            db.session.query(Post.id).filter(Post.deleted_at.isnot(None)).all()
        )

        db.session.query(Post).filter(Post.deleted_at.isnot(None)).delete(
            synchronize_session='fetch'
        )
        db.session.commit()

        return [post[0] for post in posts_to_delete]

    @classmethod
    def count(cls) -> int:
        return (
            db.session.query(func.count(Post.id))
            .filter(Post.deleted_at.is_(None))
            .scalar()
        )

    @staticmethod
    @deprecated
    def query_with_children_count() -> Query:
        # If the model does not have a `children_count` field,
        # we can dynamically calculate it using the following subquery,
        # but the performance may significantly decrease.
        subquery = (
            db.session.query(
                Post.parent_id, db.func.count(Post.id).label('children_count')
            )
            .filter(Post.deleted_at.is_(None))
            .group_by(Post.parent_id)
            .subquery()
        )
        query = (
            db.session.query(Post, subquery.c.children_count)
            .outerjoin(subquery, Post.id == subquery.c.parent_id)
            .options(subqueryload(Post.tags))  # noqa
        )
        return query

    @property
    def deleted(self) -> bool:
        return self.deleted_at is not None

    def delete(self) -> None:
        self.deleted_at = ms_now()

        if self.parent_id:
            self.parent.children_count -= 1
            db.session.add(self.parent)

        self.save()

    def restore(self) -> None:
        self.deleted_at = None

        if self.parent_id:
            self.parent.children_count += 1
            db.session.add(self.parent)

        self.save()

    def clear(self) -> None:
        db.session.delete(self)
        db.session.commit()

    def save(self) -> None:
        db.session.add(self)
        db.session.commit()
