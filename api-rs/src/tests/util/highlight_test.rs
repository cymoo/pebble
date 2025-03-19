#[cfg(test)]
mod tests {
    use pebble::util::common::highlight;

    #[test]
    fn test_empty_tokens() {
        let text = "<html><body>Hello 世界</body></html>";
        let tokens: Vec<String> = vec![];
        let result = highlight(text, &tokens);
        assert_eq!(result, text); // 应返回原始文本
    }

    #[test]
    fn test_plain_text() {
        let text = "Hello 世界";
        let tokens = vec!["Hello".to_string(), "世界".to_string()];
        let result = highlight(text, &tokens);
        assert_eq!(result, "<mark>Hello</mark> <mark>世界</mark>");
    }

    #[test]
    fn test_html_tags() {
        let text = "<html><body>Hello 世界</body></html>";
        let tokens = vec!["body".to_string()];
        let result = highlight(text, &tokens);
        assert_eq!(result, "<html><body>Hello 世界</body></html>"); // 不应替换标签内的内容
    }

    #[test]
    fn test_chinese_characters() {
        let text = "这是一个测试";
        let tokens = vec!["测试".to_string()];
        let result = highlight(text, &tokens);
        assert_eq!(result, "这是一个<mark>测试</mark>");
    }

    #[test]
    fn test_english_characters() {
        let text = "Hello world, this is a test.";
        let tokens = vec!["test".to_string(), "world".to_string()];
        let result = highlight(text, &tokens);
        assert_eq!(result, "Hello <mark>world</mark>, this is a <mark>test</mark>.");
    }

    #[test]
    fn test_mixed_content() {
        let text = "<html><body>Hello 世界, this is a 测试.</body></html>";
        let tokens = vec!["Hello".to_string(), "世界".to_string(), "test".to_string()];
        let result = highlight(text, &tokens);
        assert_eq!(
            result,
            "<html><body><mark>Hello</mark> <mark>世界</mark>, this is a 测试.</body></html>"
        );
    }

    #[test]
    fn test_repeated_tokens() {
        let text = "Hello Hello 世界 世界";
        let tokens = vec!["Hello".to_string(), "世界".to_string()];
        let result = highlight(text, &tokens);
        assert_eq!(
            result,
            "<mark>Hello</mark> <mark>Hello</mark> <mark>世界</mark> <mark>世界</mark>"
        );
    }

    #[test]
    fn test_special_characters() {
        let text = "This is a test with special characters: *test*.";
        let tokens = vec!["*test*".to_string()];
        let result = highlight(text, &tokens);
        assert_eq!(
            result,
            "This is a test with special characters: <mark>*test*</mark>."
        );
    }

    #[test]
    fn test_overlapping_tokens() {
        let text = "This is a test with overlapping tokens.";
        let tokens = vec!["test".to_string(), "with overlapping".to_string()];
        let result = highlight(text, &tokens);
        assert_eq!(
            result,
            "This is a <mark>test</mark> <mark>with overlapping</mark> tokens."
        );
    }
}
