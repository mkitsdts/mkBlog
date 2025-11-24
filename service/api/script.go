package api

const listSQL = `
		SELECT s.title, s.update_at, s.category, s.summary
		FROM article_details d
		JOIN article_summaries s ON s.title = d.title
		WHERE MATCH(d.content,d.title) AGAINST (? IN NATURAL LANGUAGE MODE)
		LIMIT ? OFFSET ?`

const countSQL = `
	SELECT COUNT(*) FROM article_details d 
	WHERE MATCH(d.content,d.title) AGAINST (? IN NATURAL LANGUAGE MODE)`

// LIKE 回退（主要用于未启用 ngram 的中文检索场景）
const listLikeSQL = `
	SELECT s.title, s.update_at, s.category, s.summary
	FROM article_details d
	JOIN article_summaries s ON s.title = d.title
	WHERE d.content LIKE ?
	ORDER BY s.update_at DESC
	LIMIT ? OFFSET ?`

const countLikeSQL = `
	SELECT COUNT(*) FROM article_details d WHERE d.content LIKE ?`
