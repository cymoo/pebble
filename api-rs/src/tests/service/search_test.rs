#[cfg(test)]
mod tests {
    use jieba_rs::Jieba;
    use mote::config::rd::RD;
    use mote::service::search_service::FullTextSearch;
    use std::sync::Arc;

    async fn setup_search() -> FullTextSearch {
        let rd = Arc::new(RD::new("redis://127.0.0.1/").await.unwrap());
        let tokenizer = Arc::new(Jieba::new());
        FullTextSearch::new(rd, tokenizer, "test:".to_string())
    }

    #[tokio::test]
    async fn test_basic_indexing_and_search() {
        let search = setup_search().await;
        search.clear_all_indexes().await.unwrap();

        // Basic indexing test
        search.index(1, "hello world").await.unwrap();
        search.index(2, "hello rust").await.unwrap();
        search.index(3, "world of rust programming").await.unwrap();

        // Simple search test
        let (tokens, results) = search.search("hello", false, 10).await.unwrap();
        assert_eq!(tokens, vec!["hello"]);
        assert_eq!(results.len(), 2);
        assert!(results.iter().any(|(id, _)| *id == 1));
        assert!(results.iter().any(|(id, _)| *id == 2));

        // Verify ranking mechanism
        let (_, results) = search.search("world", false, 10).await.unwrap();
        assert_eq!(results.len(), 2);
        assert_eq!(results[0].0, 1); // "hello world" should rank first due to shorter document length

        search.clear_all_indexes().await.unwrap();
    }

    #[tokio::test]
    async fn test_document_operations() {
        let search = setup_search().await;
        search.clear_all_indexes().await.unwrap();

        // Non-indexed document check
        assert!(!search.indexed(1).await.unwrap());

        assert_eq!(search.get_doc_count().await.unwrap(), 0);

        // Indexing a document
        search.index(1, "hello world").await.unwrap();
        assert!(search.indexed(1).await.unwrap());
        assert_eq!(search.get_doc_count().await.unwrap(), 1);

        // Reindexing the document
        search.reindex(1, "hello rust").await.unwrap();
        assert_eq!(search.get_doc_count().await.unwrap(), 1);

        // Deindexing the document
        search.deindex(1).await.unwrap();
        assert!(!search.indexed(1).await.unwrap());
        assert_eq!(search.get_doc_count().await.unwrap(), 0);

        search.clear_all_indexes().await.unwrap();
    }

    #[tokio::test]
    async fn test_index_consistency() {
        let search = setup_search().await;
        search.clear_all_indexes().await.unwrap();

        // Test index consistency when updating a document
        search.index(1, "initial content").await.unwrap();
        let initial_count = search.get_doc_count().await.unwrap();

        // Reindex the same document
        search.index(1, "updated content").await.unwrap();
        let updated_count = search.get_doc_count().await.unwrap();
        assert_eq!(initial_count, updated_count, "文档计数在更新后应保持不变");

        // Check that old content is not searchable
        let (_, results) = search.search("initial", false, 10).await.unwrap();
        assert_eq!(results.len(), 0, "旧内容不应该可被搜索到");

        // Check that new content is searchable
        let (_, results) = search.search("updated", false, 10).await.unwrap();
        assert_eq!(results.len(), 1, "新内容应该可被搜索到");

        search.clear_all_indexes().await.unwrap();
    }

    #[tokio::test]
    async fn test_partial_match() {
        let partial_search = setup_search().await;
        partial_search.clear_all_indexes().await.unwrap();

        // Prepare test data
        partial_search.index(1, "Rust Programming").await.unwrap();
        partial_search.index(2, "Python Programming").await.unwrap();
        partial_search.index(3, "Go Language").await.unwrap();

        // Test partial match search
        let (_, results) = partial_search
            .search("Rust Python", true, 10)
            .await
            .unwrap();
        assert_eq!(results.len(), 2); // Should match documents containing "rust" or "python"

        // Test full phrase match
        let (_, results) = partial_search
            .search("programming", true, 10)
            .await
            .unwrap();
        assert_eq!(results.len(), 2); // Should match all documents containing "programming"

        partial_search.clear_all_indexes().await.unwrap();
    }

    #[tokio::test]
    async fn test_edge_cases() {
        let search = setup_search().await;
        search.clear_all_indexes().await.unwrap();

        // Empty document
        search.index(1, "").await.unwrap();
        assert!(!search.indexed(1).await.unwrap());

        // Document with only spaces
        search.index(2, "   ").await.unwrap();
        assert!(!search.indexed(2).await.unwrap());

        // Document with only punctuation
        search.index(3, ".,!?;:").await.unwrap();
        assert!(!search.indexed(3).await.unwrap());

        // Document with only stop words
        search.index(4, "the and or").await.unwrap();
        assert!(!search.indexed(4).await.unwrap());

        // Very long document
        let long_text = "rust ".repeat(1000);
        search.index(5, &long_text).await.unwrap();
        let (_, results) = search.search("rust", false, 10).await.unwrap();
        assert_eq!(results.len(), 1);

        // HTML content
        search
            .index(6, "<p>Hello World</p><div>Rust</div>")
            .await
            .unwrap();
        let (_, results) = search.search("hello world rust", false, 10).await.unwrap();
        assert_eq!(results.len(), 1);

        // Special characters
        search
            .index(7, "rust#programming$language@test")
            .await
            .unwrap();
        let (_, results) = search
            .search("rust programming language test", false, 10)
            .await
            .unwrap();
        assert_eq!(results.len(), 1);

        // Unicode characters
        search
            .index(8, "rust😀programming🚀language")
            .await
            .unwrap();
        let (_, results) = search
            .search("rust programming language😀", false, 10)
            .await
            .unwrap();
        assert_eq!(results.len(), 1);

        search.clear_all_indexes().await.unwrap();
    }

    #[tokio::test]
    async fn test_cjk_support() {
        let search = setup_search().await;
        search.clear_all_indexes().await.unwrap();

        // Chinese documents
        search.index(1, "rust编程语言教程").await.unwrap();
        search.index(2, "python开发指南").await.unwrap();

        // Search with Chinese characters
        let (_, results) = search.search("编程", false, 10).await.unwrap();
        assert_eq!(results.len(), 1);
        assert_eq!(results[0].0, 1);

        // Search with mixed Chinese and English
        search
            .index(3, "学习 rust 和 python programming")
            .await
            .unwrap();
        let (_, results) = search.search("rust python", false, 10).await.unwrap();
        assert_eq!(results.len(), 1);

        // Chinese punctuation
        search
            .index(4, "rust（编程）语言，开发。教程！")
            .await
            .unwrap();
        let (_, results) = search.search("编程 语言", false, 10).await.unwrap();
        assert_eq!(results.len(), 2);

        search.clear_all_indexes().await.unwrap();
    }

    #[tokio::test]
    async fn test_ranking_and_relevance() {
        let partial_search = setup_search().await;
        partial_search.clear_all_indexes().await.unwrap();

        // Prepare documents with varying characteristics
        partial_search.index(1, "rust programming").await.unwrap();
        partial_search
            .index(2, "rust programming guide")
            .await
            .unwrap();
        partial_search
            .index(3, "rust programming complete tutorial")
            .await
            .unwrap();
        partial_search.index(4, "rust").await.unwrap();
        partial_search.index(5, "rust rust rust").await.unwrap(); // Test term frequency impact

        let (_, results) = partial_search
            .search("rust programming", true, 10)
            .await
            .unwrap();

        // Verify the ranking
        assert_eq!(results.len(), 5);
        // The most relevant document should rank first
        assert_eq!(results[0].0, 1); // Concise and contains all search terms
        assert_ne!(results[0].0, 5); // Document with repeated terms should not rank first

        partial_search.clear_all_indexes().await.unwrap();
    }

    #[tokio::test]
    async fn test_error_recovery() {
        let search = setup_search().await;
        search.clear_all_indexes().await.unwrap();

        // Attempt to deindex a non-existent document
        let result = search.deindex(999).await;
        assert!(result.is_err());
        assert!(result.unwrap_err().to_string().contains("not found"));

        // Test reindexing a non-existent document (should fallback to normal indexing)
        search.reindex(1, "test document").await.unwrap();
        assert!(search.indexed(1).await.unwrap());

        // Test repeated deletion
        search.deindex(1).await.unwrap();
        let result = search.deindex(1).await;
        assert!(result.is_err());

        // Test searching after clearing indexes
        search.clear_all_indexes().await.unwrap();
        let (_, results) = search.search("test", false, 10).await.unwrap();
        assert_eq!(results.len(), 0);
    }

    #[tokio::test]
    async fn test_max_results_limit() {
        let rd = RD::new("redis://127.0.0.1/").await.unwrap();
        let tokenizer = Arc::new(Jieba::new());
        let limited_search =
            FullTextSearch::new(Arc::new(rd), tokenizer, "test_limited:".to_string());
        limited_search.clear_all_indexes().await.unwrap();

        // Index 5 documents with the same relevance
        for i in 1..=5 {
            limited_search.index(i, "test document").await.unwrap();
        }

        // Verify result count limit
        let (_, results) = limited_search.search("test", false, 3).await.unwrap();
        assert_eq!(results.len(), 3, "结果数量应该被限制在3个");

        // Index additional documents to create varying relevance
        limited_search.index(6, "test").await.unwrap();
        limited_search.index(7, "test test test").await.unwrap();
        let (_, results) = limited_search.search("test", false, 3).await.unwrap();
        assert_eq!(results.len(), 3, "结果数量应该被限制在3个");

        limited_search.clear_all_indexes().await.unwrap();
    }
}
