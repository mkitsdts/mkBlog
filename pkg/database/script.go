package database

const (
	createMySQLFullTextIndexSQL        = "ALTER TABLE article_details ADD FULLTEXT INDEX idx_content_title_ngrams (content, title) WITH PARSER ngram;"
	createPostgresFullTextIndexSQL     = "CREATE INDEX idx_content_title_zh ON article_details USING GIN (to_tsvector('zhparser', COALESCE(content,'') || ' ' || COALESCE(title,'')));"
	usePostgresExtensionSQL            = "CREATE EXTENSION IF NOT EXISTS zhparser;"
	createPostgresChineseDictionarySQL = "CREATE TEXT SEARCH CONFIGURATION zhparser (PARSER = zhparser);"
	createPostgresDictionaryMappingSQL = "ALTER TEXT SEARCH CONFIGURATION zhparser ADD MAPPING FOR n,v,a,i,e,l WITH simple;"
)
