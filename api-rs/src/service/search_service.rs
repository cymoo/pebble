use crate::config::rd::RD;
use crate::util::common::count_frequencies;
use anyhow::{Context, Result};
use jieba_rs::Jieba;
use lazy_static::lazy_static;
use regex::Regex;
use serde::{Deserialize, Serialize};
use std::collections::{HashMap, HashSet};
use std::sync::Arc;

lazy_static! {
    static ref PUNCTUATION: Regex =
        Regex::new(r"\p{P}").expect("Failed to compile punctuation regex");
    static ref HTML_TAG: Regex = Regex::new(r"<[^>]*>").expect("Failed to compile HTML tag regex");
    static ref STOP_WORDS: HashSet<&'static str> = vec![
        "a", "an", "and", "are", "as", "at", "be", "by", "can", "for", "from", "have", "if", "in",
        "is", "it", "may", "not", "of", "on", "or", "tbd", "that", "the", "this", "to", "us", "we",
        "when", "will", "with", "yet", "you", "your", "的", "了", "和", "着", "与"
    ]
    .into_iter()
    .collect();
}

pub trait Tokenizer: Send + Sync {
    fn cut<'a>(&self, text: &'a str) -> Vec<&'a str>;

    fn analyze(&self, text: &str) -> Vec<String> {
        let text = HTML_TAG.replace_all(text, " ");

        let text = PUNCTUATION.replace_all(&text, " ");

        self.cut(&text)
            .into_iter()
            .map(str::to_lowercase)
            .filter(|token| {
                let token = token.trim();
                !token.is_empty() && !STOP_WORDS.contains(token)
            })
            .map(String::from)
            .collect()
    }
}

impl Tokenizer for Jieba {
    fn cut<'a>(&self, text: &'a str) -> Vec<&'a str> {
        self.cut_for_search(text, false)
    }
}

#[derive(Debug, Serialize, Deserialize)]
struct TokenFrequency(HashMap<String, usize>);

pub struct FullTextSearch {
    rd: Arc<RD>,
    tokenizer: Arc<dyn Tokenizer>,
    key_prefix: String,
}

impl FullTextSearch {
    pub fn new(rd: Arc<RD>, tokenizer: Arc<dyn Tokenizer>, key_prefix: String) -> Self {
        Self {
            rd,
            tokenizer,
            key_prefix,
        }
    }

    pub async fn indexed(&self, id: i64) -> Result<bool> {
        self.rd.exists(self.doc_tokens_key(id)).await
    }

    pub async fn get_doc_count(&self) -> Result<i64> {
        let count: Option<String> = self.rd.get(self.doc_count_key()).await?;
        count
            .unwrap_or("0".to_string())
            .parse::<i64>()
            .context("Failed to parse doc count")
    }

    pub async fn index(&self, id: i64, text: &str) -> Result<()> {
        if self.indexed(id).await? {
            // a recursive async fn call must introduce indirection,
            // such as Box::pin to avoid an infinitely sized future
            return Box::pin(self.reindex(id, text)).await;
        }

        let tokens = self.tokenizer.analyze(text);
        if tokens.is_empty() {
            return Ok(());
        }

        let token_frequency = count_frequencies(&tokens);
        let freq_json = serde_json::to_string(&TokenFrequency(token_frequency))?;

        let token_set = tokens.into_iter().collect::<HashSet<String>>();

        let _: () = self
            .rd
            .pipeline(|pipe| {
                pipe.set(self.doc_tokens_key(id), freq_json);
                pipe.incr(self.doc_count_key(), 1);
                for token in token_set.iter() {
                    pipe.sadd(self.token_docs_key(token), id);
                }
            })
            .await?;

        Ok(())
    }

    pub async fn reindex(&self, id: i64, text: &str) -> Result<()> {
        if !self.indexed(id).await? {
            return Box::pin(self.index(id, text)).await;
        }

        let new_tokens = self.tokenizer.analyze(text);
        if new_tokens.is_empty() {
            return self.deindex(id).await;
        }

        let old_freq: TokenFrequency = self
            .rd
            .get_object(self.doc_tokens_key(id))
            .await?
            .ok_or(anyhow::anyhow!("Token frequency of doc `{}` not found", id))?;

        let new_freq = count_frequencies(&new_tokens);
        let freq_json = serde_json::to_string(&TokenFrequency(new_freq))?;

        let old_token_set = old_freq.0.keys().collect::<HashSet<_>>();
        let new_token_set = new_tokens.iter().collect::<HashSet<_>>();
        let tokens_to_remove = old_token_set.difference(&new_token_set).collect::<Vec<_>>();
        let tokens_to_add = new_token_set.difference(&old_token_set).collect::<Vec<_>>();

        let _: () = self
            .rd
            .pipeline(|pipe| {
                pipe.set(self.doc_tokens_key(id), freq_json);
                for token in tokens_to_remove {
                    pipe.srem(self.token_docs_key(token), id);
                }
                for token in tokens_to_add {
                    pipe.sadd(self.token_docs_key(token), id);
                }
            })
            .await?;

        Ok(())
    }

    pub async fn deindex(&self, id: i64) -> Result<()> {
        let token_freq: TokenFrequency = self
            .rd
            .get_object(self.doc_tokens_key(id))
            .await?
            .ok_or(anyhow::anyhow!("Token frequency of doc `{}` not found", id))?;

        let token_set = token_freq.0.keys().collect::<HashSet<_>>();

        let _: () = self
            .rd
            .pipeline(|pipe| {
                pipe.del(self.doc_tokens_key(id));
                pipe.decr(self.doc_count_key(), 1);
                for token in token_set.iter() {
                    pipe.srem(self.token_docs_key(token), id);
                }
            })
            .await?;

        Ok(())
    }

    pub async fn search(
        &self,
        query: &str,
        partial: bool,
        limit: usize,
    ) -> Result<(Vec<String>, Vec<(i64, f64)>)> {
        let tokens = self.tokenizer.analyze(query);
        if tokens.is_empty() {
            return Ok((tokens, vec![]));
        }

        // Retrieve the document IDs containing the query term
        let doc_sets: Vec<HashSet<String>> = self
            .rd
            .pipeline(|pipe| {
                for token in tokens.iter() {
                    pipe.smembers(self.token_docs_key(token));
                }
            })
            .await?;

        let ids: HashSet<i64> = if partial {
            // union
            doc_sets
                .into_iter()
                .flatten()
                .filter_map(|id| id.parse().ok())
                .collect()
        } else {
            // intersection
            doc_sets
                .into_iter()
                .reduce(|acc, set| acc.intersection(&set).cloned().collect())
                .unwrap_or_default()
                .into_iter()
                .filter_map(|id| id.parse().ok())
                .collect()
        };

        if ids.is_empty() {
            return Ok((tokens, vec![]));
        }

        // Calculate the relevance score
        let mut ranked_results = self.rank(&tokens, &ids).await?;
        ranked_results.sort_by(|a, b| b.1.partial_cmp(&a.1).unwrap());

        // Limit the number of results
        if limit > 0 && ranked_results.len() > limit {
            ranked_results.truncate(limit);
        }

        Ok((tokens, ranked_results))
    }

    async fn rank(&self, tokens: &[String], ids: &HashSet<i64>) -> Result<Vec<(i64, f64)>> {
        let mut results = Vec::new();

        let total_docs = self.get_doc_count().await? as f64;

        let token_frequencies: Vec<Option<TokenFrequency>> = self
            .rd
            .mget_object(
                ids.iter()
                    .map(|id| self.doc_tokens_key(*id))
                    .collect::<Vec<String>>(),
            )
            .await?;

        let doc_frequencies: Vec<f64> = self
            .rd
            .pipeline(|pipe| {
                for token in tokens.iter() {
                    pipe.scard(self.token_docs_key(token));
                }
            })
            .await?;

        for (&id, token_frequency) in ids.iter().zip(token_frequencies.iter()) {
            let token_freq = token_frequency
                .as_ref()
                .expect(&format!("Token frequency of doc `{}` not found", id));

            let mut score = 0.0;
            let mut matching_terms = 0;

            for (token, df) in tokens.iter().zip(doc_frequencies.iter()) {
                let tf = *token_freq.0.get(token).unwrap_or(&0) as f64;
                if tf > 0.0 {
                    matching_terms += 1;
                }

                // Use an improved TF calculation: 1 + log(tf) to reduce the weight of high-frequency terms
                let normalized_tf = if tf > 0.0 { 1.0 + (tf.log10()) } else { 0.0 };

                let idf = if *df > 0.0 {
                    (total_docs / df).max(1.0).log10()
                } else {
                    0.0
                };

                score += normalized_tf * idf;
            }

            // Apply a length normalization factor to avoid advantages for long documents
            let total_terms = token_freq.0.values().sum::<usize>() as f64;
            if total_terms > 0.0 {
                score /= total_terms.sqrt();
            }

            // Calculate query term coverage
            let coverage_ratio = matching_terms as f64 / tokens.len() as f64;
            score *= if coverage_ratio > 0.999 {
                2.0
            } else {
                coverage_ratio
            };

            results.push((id, score));
        }

        Ok(results)
    }

    fn doc_count_key(&self) -> String {
        format!("{}doc:count", self.key_prefix)
    }

    fn doc_tokens_key(&self, id: i64) -> String {
        format!("{}doc:{}:tokens", self.key_prefix, id)
    }

    fn token_docs_key(&self, token: &str) -> String {
        format!("{}token:{}:docs", self.key_prefix, token)
    }

    pub async fn clear_all_indexes(&self) -> Result<()> {
        let prefixes = [
            format!("{}doc:", self.key_prefix),
            format!("{}token:", self.key_prefix),
        ];

        for prefix in prefixes.iter() {
            let keys: Vec<String> = self.rd.keys(format!("{}*", prefix)).await?;
            if !keys.is_empty() {
                self.rd.del(&keys).await?;
            }
        }

        Ok(())
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    async fn setup() -> FullTextSearch {
        let rd = Arc::new(RD::new("redis://127.0.0.1").await.unwrap());
        let tokenizer = Arc::new(Jieba::new());
        FullTextSearch::new(rd, tokenizer, "test:".to_owned())
    }

    #[tokio::test]
    async fn smoke_test() {
        let fts = setup().await;

        fts.index(1, "测试文档 hello world").await.unwrap();
        assert!(fts.indexed(1).await.unwrap());

        assert_eq!(fts.get_doc_count().await.unwrap(), 1);

        let (_, results) = fts.search("hello", true, 300).await.unwrap();
        assert_eq!(results.len(), 1);

        let (_, results) = fts.search("测试", true, 300).await.unwrap();
        assert_eq!(results.len(), 1);

        let (_, results) = fts.search("hello rust", true, 300).await.unwrap();
        assert_eq!(results.len(), 1);

        fts.clear_all_indexes().await.unwrap();

        assert_eq!(fts.get_doc_count().await.unwrap(), 0);
    }
}
