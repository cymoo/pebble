from app.util import highlight_html


def test_mark_tokens():
    # Test case 1: Basic text marking
    assert highlight_html(['hello'], 'hello world') == '<mark>hello</mark> world'

    # Test case 2: Multiple tokens
    assert (
        highlight_html(['hello', 'world'], 'hello world')
        == '<mark>hello</mark> <mark>world</mark>'
    )

    # Test case 3: Chinese text
    assert highlight_html(['你好'], '你好世界') == '<mark>你好</mark>世界'

    # Test case 4: Inside HTML tags
    assert (
        highlight_html(['token'], '<a href="token">token</a>')
        == '<a href="token"><mark>token</mark></a>'
    )

    # Test case 5: Multiple attributes
    assert (
        highlight_html(['test'], '<div class="test" data-test="test">test</div>')
        == '<div class="test" data-test="test"><mark>test</mark></div>'
    )

    # Test case 6: Text with < and >
    # TODO: this case will fail
    # assert (
    #     mark_tokens(['token'], '<p>some token >~< token</p>')
    #     == '<p>some <mark>token</mark> >~< <mark>token</mark></p>'
    # )

    # Test case 7: Empty token list
    assert highlight_html([], '<p>test</p>') == '<p>test</p>'

    # Test case 8: Overlapping tokens
    assert (
        highlight_html(['hello world', 'world'], 'hello world')
        == '<mark>hello world</mark>'
    )

    # Test case 9: Special characters
    assert (
        highlight_html(['test.com', 'test'], '<a href="test.com">test.com test</a>')
        == '<a href="test.com"><mark>test.com</mark> <mark>test</mark></a>'
    )

    # Test case 10: Mixed content
    assert (
        highlight_html(
            ['hello', '你好', 'world', '世界'], '<div>hello 世界, 你好 world!</div>'
        )
        == '<div><mark>hello</mark> <mark>世界</mark>, <mark>你好</mark> <mark>world</mark>!</div>'
    )

    # Test case 11: English word boundary test
    assert (
        highlight_html(['foo'], 'This is foo and foolish')
        == 'This is <mark>foo</mark> and foolish'
    )

    # Test case 11: English word boundary and mixed content
    assert (
        highlight_html(['foo', '你好'], 'This is foo and foolish, 你好 world')
        == 'This is <mark>foo</mark> and foolish, <mark>你好</mark> world'
    )
