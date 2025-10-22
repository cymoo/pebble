import os
from concurrent.futures import ThreadPoolExecutor
from mimetypes import guess_extension
from os import path

from PIL import Image, ImageOps
from flask import (
    Blueprint,
    request,
    Response,
    abort,
    current_app as app,
    url_for,
    stream_with_context,
)

from ..dto import *
from ..exception import bad_request
from ..model import Post, Tag, HASH_PATTERN
from ..extension import fts, db
from ..util import (
    gen_thumbnail,
    add_suffix,
    gen_secure_filename,
    mark_tokens_in_html,
    parse_date_with_timezone,
)
from ..middleware import rate_limit, validate

api = Blueprint('api', __name__)

executor = ThreadPoolExecutor()


def is_valid_password(password: str) -> bool:
    return password == os.getenv('PEBBLE_PASSWORD')


@api.get('/')
def index():
    return {'msg': 'hello world'}


@api.post('/login')
@rate_limit(max_count=10, expires=60)
@validate
def login(payload: LoginRequest) -> NoContent:
    password = payload.password
    if not is_valid_password(password):
        bad_request('wrong password')
    return NO_CONTENT


@api.get('/get-tags')
def get_tags() -> list[TagDto]:
    return [
        TagDto(name=name, sticky=sticky, post_count=count)
        for (name, sticky, count) in Tag.get_all_with_post_count()
    ]


@api.post('/stick-tag')
@validate
def stick_tag(payload: StickTagRequest) -> NoContent:
    Tag.insert_or_update(payload.name, payload.sticky)
    return NO_CONTENT


@api.post('/rename-tag')
@rate_limit(max_count=5, expires=60)
@validate
def rename_tag(payload: RenameTagRequest) -> NoContent:
    Tag.rename_or_merge(payload.name, payload.new_name)
    return NO_CONTENT


@api.post('/delete-tag')
@rate_limit(max_count=1, expires=10)
@validate
def delete_tag(payload: Name) -> NoContent:
    tag = Tag.find_by_name(payload.name)
    if not tag:
        bad_request('tag not found')
    tag.delete()

    return NO_CONTENT


@api.get('/search')
@validate(type='query')
def search(payload: SearchRequest) -> PostPagination:
    query, partial, limit = payload.query, payload.partial, payload.limit

    tokens, results = fts.search(query, partial, limit)

    if not results:
        return PostPagination(posts=[], cursor=-1, size=0)

    scores = {id: score for id, score in results}
    posts = Post.find_by_ids(scores.keys())

    posts_with_score = []
    for post in posts:  # noqa
        post = PostDto.from_model(post)
        post.content = mark_tokens_in_html(tokens, post.content)
        post.score = scores[post.id]
        posts_with_score.append(post)

    posts_with_score.sort(key=lambda x: (x.score, x.created_at), reverse=True)

    return PostPagination(posts=posts_with_score, cursor=-1, size=len(posts_with_score))


@api.get('/get-posts')
@validate(type='query')
def get_posts(payload: FilterPostRequest) -> PostPagination:
    posts = Post.filter_posts(
        **payload.model_dump(),
        per_page=app.config['POSTS_PER_PAGE'],
    )
    posts = [PostDto.from_model(post) for post in posts]

    return PostPagination(
        posts=posts, cursor=posts[-1].created_at if posts else -1, size=len(posts)
    )


@api.get('/get-post')
@validate(type='query')
def get_post(payload: Id) -> PostDto:
    post = db.get_or_404(Post, payload.id, description='post not found')
    if post.deleted:
        abort(404, 'post not found')

    return PostDto.from_model(post)


@api.post('/create-post')
@validate
def create_post(payload: CreatePostRequest) -> CreationDto:
    post = Post(
        content=payload.content,
        files=json.dumps(payload.files) if payload.files else None,
        color=payload.color,
        shared=payload.shared or False,
        parent_id=payload.parent_id,
    )
    post.save()

    executor.submit(fts.index, post.id, payload.content)

    return CreationDto(
        id=post.id,
        created_at=post.created_at,
        updated_at=post.updated_at,
    )


@api.post('/update-post')
@validate
def update_post(payload: UpdatePostRequest) -> NoContent:
    post = db.get_or_404(Post, payload.id, description='post not found')
    if post.deleted:
        abort(404)

    old_content = post.content

    for field, value in payload.model_dump().items():
        if value is missing:
            continue

        if field == 'content':
            post.content = value
            post.tags = [Tag.find_or_create(tag) for tag in HASH_PATTERN.findall(value)]
        elif field == 'files':
            post.files = json.dumps(value) if value else None
        elif field == 'parent_id':
            if post.parent_id is not None and value is None:
                post.parent.children_count -= 1
            elif post.parent_id is None and value is not None:
                post.parent.children_count += 1
            post.parent_id = value
        else:
            setattr(post, field, value)

    post.save()

    if payload.content is not None and post.content != old_content:
        executor.submit(fts.reindex, post.id, payload.content)

    return NO_CONTENT


@api.post('/delete-post')
@validate
def delete_post(payload: DeletePostRequest) -> NoContent:
    hard, id = payload.hard, payload.id  # noqa
    post = db.get_or_404(Post, id, description='post not found')

    if hard:
        post.clear()
        executor.submit(fts.deindex, id)
    else:
        post.delete()

    return NO_CONTENT


@api.post('/restore-post')
@validate
def restore_post(payload: Id) -> NoContent:
    post = db.get_or_404(Post, payload.id, description='post not found')
    post.restore()

    return NO_CONTENT


@api.post('/clear-posts')
def clear_posts() -> NoContent:
    ids = Post.clear_all()

    for _id in ids:
        executor.submit(fts.deindex, _id)
    return NO_CONTENT


@api.get('/get-daily-post-counts')
@validate(type='query')
def get_daily_post_counts(payload: DateRange) -> list[int]:
    return Post.get_daily_counts(
        start_date=parse_date_with_timezone(payload.start_date, payload.offset),
        end_date=parse_date_with_timezone(
            payload.end_date, payload.offset, at_end=True
        ),
    )


@api.get('/get-overall-counts')
def get_stats() -> PostStats:
    return PostStats(
        post_count=Post.count(),
        tag_count=Tag.count(),
        day_count=Post.get_active_days(),
    )


# For quick test
@api.get('/upload')
def file_form() -> str:
    return """
    <!doctype html>
    <html>
        <head><title>Upload file</title></head>
        <body>
            <form action="upload" method="post" enctype="multipart/form-data">
                <input type="file" name="file" multiple>
                <button type="submit">Upload</button>
            </form>
        </body>
    </html>
    """.strip()


@api.post('/upload')
def upload_file() -> FileInfo:
    config = app.config

    file = request.files.get('file')
    if not file:
        bad_request('no uploaded file')
    if not file.filename:
        bad_request('filename is required')

    filename = gen_secure_filename(file.filename)
    filepath = path.join(config['UPLOAD_PATH'], filename)

    ext = guess_extension(file.mimetype)

    if ext not in config['UPLOAD_IMAGE_FORMATS']:
        file.save(filepath)
        return FileInfo(
            url=url_for('uploaded_file', filename=filename),
            # NOTE: file.content_length is not reliable: https://stackoverflow.com/questions/15772975
            size=os.stat(filepath).st_size,
        )
    else:
        with Image.open(file.stream) as img:
            thumb_url = None
            if ext != '.gif':
                # https://stackoverflow.com/questions/13872331/rotating-an-image-with-orientation-specified-in-exif-using-python-without-pil-in
                img = ImageOps.exif_transpose(img)
                # TODO: sometimes thumbnail generation will fail, figure out why later
                image_thumb = gen_thumbnail(img, config['UPLOAD_THUMB_WIDTH'])
                filename_thumb = add_suffix(filename, '.thumb')
                image_thumb.save(
                    path.join(config['UPLOAD_PATH'], filename_thumb), quality=90
                )
                thumb_url = url_for('uploaded_file', filename=filename_thumb)

            # NOTE: Perhaps the image should be compressed
            # why `save_all`: https://stackoverflow.com/questions/24688802/saving-an-animated-gif-in-pillow
            img.save(filepath, save_all=ext == '.gif')

            return FileInfo(
                url=url_for('uploaded_file', filename=filename),
                size=os.stat(filepath).st_size,
                width=img.width,
                height=img.height,
                thumb_url=thumb_url,
            )


@api.get('/_dangerously_rebuild_all_indexes')
@rate_limit(max_count=3, expires=60 * 60)
def rebuild_indexes() -> Response:
    posts = Post.query.with_entities(Post.id, Post.content).all()

    def generate():
        yield 'Indexing...\n'
        fts.clear_all_indexes()
        for post in posts:
            fts.index(post.id, post.content)
        yield 'Done'

    return Response(stream_with_context(generate()), content_type='text/plain')


@api.get('/auth')
def auth() -> NoContent:
    return NO_CONTENT


@api.before_request
def check_permission() -> None:
    if request.path.endswith('login'):
        return

    token = request.cookies.get('token')

    if not token:
        auth_header = request.headers.get('Authorization')
        if not auth_header:
            abort(401, "missing authorization header")
        if not auth_header.startswith('Bearer '):
            abort(401, "invalid authorization header")

        token = auth_header[len('Bearer ') :].strip()

    if not is_valid_password(token):
        abort(401, "invalid authorization token")
