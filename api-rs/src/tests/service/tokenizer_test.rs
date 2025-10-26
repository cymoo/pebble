#[cfg(test)]
mod tests {
    use jieba_rs::Jieba;
    use pebble::service::search_service::Tokenizer;

    fn init_tokenizer() -> Jieba {
        Jieba::new()
    }

    #[test]
    fn test_html_tag_removal() {
        let tokenizer = init_tokenizer();
        let text = "这是<p>一段</p><b>HTML</b>文本";
        let result = tokenizer.analyze(text);
        assert!(!result.contains(&String::from("<p>")));
        assert!(!result.contains(&String::from("</p>")));
        assert!(!result.contains(&String::from("<b>")));
        assert!(!result.contains(&String::from("</b>")));
    }

    #[test]
    fn test_punctuation_removal() {
        let tokenizer = init_tokenizer();
        let text = "Hello, world! How are you?";
        let result = tokenizer.analyze(text);
        assert!(!result.contains(&String::from(",")));
        assert!(!result.contains(&String::from("!")));
        assert!(!result.contains(&String::from("?")));
    }

    #[test]
    fn test_stop_words_removal() {
        let tokenizer = init_tokenizer();
        let text = "this is a test text";
        let result = tokenizer.analyze(text);
        assert!(!result.contains(&String::from("this")));
        assert!(!result.contains(&String::from("is")));
        assert!(!result.contains(&String::from("a")));
    }

    #[test]
    fn test_chinese_stop_words_removal() {
        let tokenizer = init_tokenizer();
        let text = "我们的世界和未来";
        let result = tokenizer.analyze(text);
        assert!(!result.contains(&String::from("的")));
        assert!(!result.contains(&String::from("和")));
    }

    #[test]
    fn test_case_normalization() {
        let tokenizer = init_tokenizer();
        let text = "Hello WORLD";
        let result = tokenizer.analyze(text);
        assert!(result.contains(&String::from("hello")));
        assert!(result.contains(&String::from("world")));
        assert!(!result.contains(&String::from("Hello")));
        assert!(!result.contains(&String::from("WORLD")));
    }

    #[test]
    fn test_mixed_content() {
        let tokenizer = init_tokenizer();
        let text = "<div>Hello, 世界! The quick brown fox.</div>";
        let result = tokenizer.analyze(text);

        // HTML tags should be removed
        assert!(!result.contains(&String::from("<div>")));
        assert!(!result.contains(&String::from("</div>")));

        // Punctuation should be removed
        assert!(!result.contains(&String::from(",")));
        assert!(!result.contains(&String::from("!")));
        assert!(!result.contains(&String::from(".")));

        // Stop words should be removed
        assert!(!result.contains(&String::from("the")));

        // Should contain the meaningful words
        assert!(result.contains(&String::from("hello")));
        assert!(result.contains(&String::from("世界")));
        assert!(result.contains(&String::from("quick")));
        assert!(result.contains(&String::from("brown")));
        assert!(result.contains(&String::from("fox")));
    }

    #[test]
    fn test_empty_text() {
        let tokenizer = init_tokenizer();
        let text = "";
        let result = tokenizer.analyze(text);
        assert!(result.is_empty());
    }

    #[test]
    fn test_only_stop_words() {
        let tokenizer = init_tokenizer();
        let text = "the a an and";
        let result = tokenizer.analyze(text);
        assert!(result.is_empty());
    }

    #[test]
    fn test_only_html_tags() {
        let tokenizer = init_tokenizer();
        let text = "<p></p><div></div><span></span>";
        let result = tokenizer.analyze(text);
        assert!(result.is_empty());
    }

    #[test]
    fn test_only_punctuation() {
        let tokenizer = init_tokenizer();
        let text = ",.!?;:\"'";
        let result = tokenizer.analyze(text);
        assert!(result.is_empty());
    }
}
