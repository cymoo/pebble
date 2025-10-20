#[cfg(test)]
mod tests {
    use jieba_rs::Jieba;
    use pebble::config::rd::RD;
    use pebble::service::search_service::FullTextSearch;
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

        // 测试基本索引
        search.index(1, "hello world").await.unwrap();
        search.index(2, "hello rust").await.unwrap();
        search.index(3, "world of rust programming").await.unwrap();

        // 测试简单搜索
        let (tokens, results) = search.search("hello", false, 10).await.unwrap();
        assert_eq!(tokens, vec!["hello"]);
        assert_eq!(results.len(), 2);
        assert!(results.iter().any(|(id, _)| *id == 1));
        assert!(results.iter().any(|(id, _)| *id == 2));

        // 验证评分机制
        let (_, results) = search.search("world", false, 10).await.unwrap();
        assert_eq!(results.len(), 2);
        assert_eq!(results[0].0, 1); // "hello world" 应该排在前面因为文档更短

        search.clear_all_indexes().await.unwrap();
    }

    #[tokio::test]
    async fn test_document_operations() {
        let search = setup_search().await;
        search.clear_all_indexes().await.unwrap();

        // 测试索引存在性检查
        assert!(!search.indexed(1).await.unwrap());

        // 测试文档计数
        assert_eq!(search.get_doc_count().await.unwrap(), 0);

        // 测试索引
        search.index(1, "hello world").await.unwrap();
        assert!(search.indexed(1).await.unwrap());
        assert_eq!(search.get_doc_count().await.unwrap(), 1);

        // 测试重新索引
        search.reindex(1, "hello rust").await.unwrap();
        assert_eq!(search.get_doc_count().await.unwrap(), 1);

        // 测试删除索引
        search.deindex(1).await.unwrap();
        assert!(!search.indexed(1).await.unwrap());
        assert_eq!(search.get_doc_count().await.unwrap(), 0);

        search.clear_all_indexes().await.unwrap();
    }

    #[tokio::test]
    async fn test_index_consistency() {
        let search = setup_search().await;
        search.clear_all_indexes().await.unwrap();

        // 测试更新文档时的索引一致性
        search.index(1, "initial content").await.unwrap();
        let initial_count = search.get_doc_count().await.unwrap();

        // 重新索引同一文档
        search.index(1, "updated content").await.unwrap();
        let updated_count = search.get_doc_count().await.unwrap();
        assert_eq!(initial_count, updated_count, "文档计数在更新后应保持不变");

        // 验证旧内容不可搜索
        let (_, results) = search.search("initial", false, 10).await.unwrap();
        assert_eq!(results.len(), 0, "旧内容不应该可被搜索到");

        // 验证新内容可搜索
        let (_, results) = search.search("updated", false, 10).await.unwrap();
        assert_eq!(results.len(), 1, "新内容应该可被搜索到");

        search.clear_all_indexes().await.unwrap();
    }

    #[tokio::test]
    async fn test_partial_match() {
        let partial_search = setup_search().await;
        partial_search.clear_all_indexes().await.unwrap();

        // 准备测试数据
        partial_search.index(1, "Rust Programming").await.unwrap();
        partial_search.index(2, "Python Programming").await.unwrap();
        partial_search.index(3, "Go Language").await.unwrap();

        // 测试部分匹配搜索
        let (_, results) = partial_search
            .search("Rust Python", true, 10)
            .await
            .unwrap();
        assert_eq!(results.len(), 2); // 应该匹配包含 "rust" 或 "python" 的文档

        // 测试完整词组匹配
        let (_, results) = partial_search
            .search("programming", true, 10)
            .await
            .unwrap();
        assert_eq!(results.len(), 2); // 应该匹配所有包含 "programming" 的文档

        partial_search.clear_all_indexes().await.unwrap();
    }

    #[tokio::test]
    async fn test_edge_cases() {
        let search = setup_search().await;
        search.clear_all_indexes().await.unwrap();

        // 空文档测试
        search.index(1, "").await.unwrap();
        assert!(!search.indexed(1).await.unwrap());

        // 只有空格的文档
        search.index(2, "   ").await.unwrap();
        assert!(!search.indexed(2).await.unwrap());

        // 只有标点符号的文档
        search.index(3, ".,!?;:").await.unwrap();
        assert!(!search.indexed(3).await.unwrap());

        // 只有停用词的文档
        search.index(4, "the and or").await.unwrap();
        assert!(!search.indexed(4).await.unwrap());

        // 超长文档
        let long_text = "rust ".repeat(1000);
        search.index(5, &long_text).await.unwrap();
        let (_, results) = search.search("rust", false, 10).await.unwrap();
        assert_eq!(results.len(), 1);

        // HTML内容
        search
            .index(6, "<p>Hello World</p><div>Rust</div>")
            .await
            .unwrap();
        let (_, results) = search.search("hello world rust", false, 10).await.unwrap();
        assert_eq!(results.len(), 1);

        // 特殊字符
        search
            .index(7, "rust#programming$language@test")
            .await
            .unwrap();
        let (_, results) = search
            .search("rust programming language test", false, 10)
            .await
            .unwrap();
        assert_eq!(results.len(), 1);

        // Unicode字符
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

        // 中文文档
        search.index(1, "rust编程语言教程").await.unwrap();
        search.index(2, "python开发指南").await.unwrap();

        // 中文搜索
        let (_, results) = search.search("编程", false, 10).await.unwrap();
        assert_eq!(results.len(), 1);
        assert_eq!(results[0].0, 1);

        // 混合语言文档
        search
            .index(3, "学习 rust 和 python programming")
            .await
            .unwrap();
        let (_, results) = search.search("rust python", false, 10).await.unwrap();
        assert_eq!(results.len(), 1);

        // 中文标点符号
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

        // 准备具有不同特征的文档
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
        partial_search.index(5, "rust rust rust").await.unwrap(); // 测试词频影响

        let (_, results) = partial_search
            .search("rust programming", true, 10)
            .await
            .unwrap();

        // 验证评分
        assert_eq!(results.len(), 5);
        // 最匹配的文档应该排在前面
        assert_eq!(results[0].0, 1); // 简短且包含所有搜索词
                                     // 重复词的文档不应该排在最前面
        assert_ne!(results[0].0, 5);

        partial_search.clear_all_indexes().await.unwrap();
    }

    #[tokio::test]
    async fn test_error_recovery() {
        let search = setup_search().await;
        search.clear_all_indexes().await.unwrap();

        // 测试删除不存在的文档
        let result = search.deindex(999).await;
        assert!(result.is_err());
        assert!(result.unwrap_err().to_string().contains("not found"));

        // 测试重新索引不存在的文档（应该回退到普通索引）
        search.reindex(1, "test document").await.unwrap();
        assert!(search.indexed(1).await.unwrap());

        // 测试重复删除
        search.deindex(1).await.unwrap();
        let result = search.deindex(1).await;
        assert!(result.is_err());

        // 测试在清理后搜索
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

        // 索引5个相同相关度的文档
        for i in 1..=5 {
            limited_search.index(i, "test document").await.unwrap();
        }

        // 验证结果数量限制
        let (_, results) = limited_search.search("test", false, 3).await.unwrap();
        assert_eq!(results.len(), 3, "结果数量应该被限制在3个");

        // 验证不同相关度的情况
        limited_search.index(6, "test").await.unwrap();
        limited_search.index(7, "test test test").await.unwrap();
        let (_, results) = limited_search.search("test", false, 3).await.unwrap();
        assert_eq!(results.len(), 3, "结果数量应该被限制在3个");

        limited_search.clear_all_indexes().await.unwrap();
    }
}
