package database

const createNgramFullTextIndexSQL = `
	ALTER TABLE article_details
	ADD FULLTEXT INDEX idx_content_title_ngrams (content, title) WITH PARSER ngram;
`
