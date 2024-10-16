package sqlfmt

var (
	n1qlReservedWords = []string{
		"ALL",
		"ALTER",
		"ANALYZE",
		"AND",
		"ANY",
		"ARRAY",
		"AS",
		"ASC",
		"BEGIN",
		"BETWEEN",
		"BINARY",
		"BOOLEAN",
		"BREAK",
		"BUCKET",
		"BUILD",
		"BY",
		"CALL",
		"CASE",
		"CAST",
		"CLUSTER",
		"COLLATE",
		"COLLECTION",
		"COMMIT",
		"CONNECT",
		"CONTINUE",
		"CORRELATE",
		"COVER",
		"CREATE",
		"DATABASE",
		"DATASET",
		"DATASTORE",
		"DECLARE",
		"DECREMENT",
		"DELETE",
		"DERIVED",
		"DESC",
		"DESCRIBE",
		"DISTINCT",
		"DO",
		"DROP",
		"EACH",
		"ELEMENT",
		"ELSE",
		"END",
		"EVERY",
		"EXCEPT",
		"EXCLUDE",
		"EXECUTE",
		"EXISTS",
		"EXPLAIN",
		"FALSE",
		"FETCH",
		"FIRST",
		"FLATTEN",
		"FOR",
		"FORCE",
		"FROM",
		"FUNCTION",
		"GRANT",
		"GROUP",
		"GSI",
		"HAVING",
		"IF",
		"IGNORE",
		"ILIKE",
		"IN",
		"INCLUDE",
		"INCREMENT",
		"INDEX",
		"INFER",
		"INLINE",
		"INNER",
		"INSERT",
		"INTERSECT",
		"INTO",
		"IS",
		"JOIN",
		"KEY",
		"KEYS",
		"KEYSPACE",
		"KNOWN",
		"LAST",
		"LEFT",
		"LET",
		"LETTING",
		"LIKE",
		"LIMIT",
		"LSM",
		"MAP",
		"MAPPING",
		"MATCHED",
		"MATERIALIZED",
		"MERGE",
		"MISSING",
		"NAMESPACE",
		"NEST",
		"NOT",
		"NULL",
		"NUMBER",
		"OBJECT",
		"OFFSET",
		"ON",
		"OPTION",
		"OR",
		"ORDER",
		"OUTER",
		"OVER",
		"PARSE",
		"PARTITION",
		"PASSWORD",
		"PATH",
		"POOL",
		"PREPARE",
		"PRIMARY",
		"PRIVATE",
		"PRIVILEGE",
		"PROCEDURE",
		"PUBLIC",
		"RAW",
		"REALM",
		"REDUCE",
		"RENAME",
		"RETURN",
		"RETURNING",
		"REVOKE",
		"RIGHT",
		"ROLE",
		"ROLLBACK",
		"SATISFIES",
		"SCHEMA",
		"SELECT",
		"SELF",
		"SEMI",
		"SET",
		"SHOW",
		"SOME",
		"START",
		"STATISTICS",
		"STRING",
		"SYSTEM",
		"THEN",
		"TO",
		"TRANSACTION",
		"TRIGGER",
		"TRUE",
		"TRUNCATE",
		"UNDER",
		"UNION",
		"UNIQUE",
		"UNKNOWN",
		"UNNEST",
		"UNSET",
		"UPDATE",
		"UPSERT",
		"USE",
		"USER",
		"USING",
		"VALIDATE",
		"VALUE",
		"VALUED",
		"VALUES",
		"VIA",
		"VIEW",
		"WHEN",
		"WHERE",
		"WHILE",
		"WITH",
		"WITHIN",
		"WORK",
		"XOR",
	}

	n1qlReservedTopLevelWords = []string{
		"DELETE FROM",
		"EXCEPT ALL",
		"EXCEPT",
		"EXPLAIN DELETE FROM",
		"EXPLAIN UPDATE",
		"EXPLAIN UPSERT",
		"FROM",
		"GROUP BY",
		"HAVING",
		"INFER",
		"INSERT INTO",
		"LET",
		"LIMIT",
		"MERGE",
		"NEST",
		"ORDER BY",
		"PREPARE",
		"SELECT",
		"SET CURRENT SCHEMA",
		"SET SCHEMA",
		"SET",
		"UNNEST",
		"UPDATE",
		"UPSERT",
		"USE KEYS",
		"VALUES",
		"WHERE",
	}

	n1qlReservedTopLevelWordsNoIndent = []string{"INTERSECT", "INTERSECT ALL", "MINUS", "UNION", "UNION ALL"}

	n1qlReservedNewlineWords = []string{
		"AND",
		"INNER JOIN",
		"JOIN",
		"LEFT JOIN",
		"LEFT OUTER JOIN",
		"OR",
		"OUTER JOIN",
		"RIGHT JOIN",
		"RIGHT OUTER JOIN",
		"XOR",
	}
)

type N1QLFormatter struct {
	cfg *Config
}

func NewN1QLFormatter(cfg *Config) *N1QLFormatter {
	cfg.TokenizerConfig = NewN1QLTokenizerConfig()
	return &N1QLFormatter{cfg: cfg}
}

func NewN1QLTokenizerConfig() *TokenizerConfig {
	return &TokenizerConfig{
		ReservedWords:                 n1qlReservedWords,
		ReservedTopLevelWords:         n1qlReservedTopLevelWords,
		ReservedNewlineWords:          n1qlReservedNewlineWords,
		ReservedTopLevelWordsNoIndent: n1qlReservedTopLevelWordsNoIndent,
		StringTypes:                   []string{`""`, "''", "``", "$$"},
		OpenParens:                    []string{"(", "[", "{"},
		CloseParens:                   []string{")", "]", "}"},
		NamedPlaceholderTypes:         []string{"$"},
		LineCommentTypes:              []string{"--", "#"},
	}
}

func (ssf *N1QLFormatter) Format(query string) string {
	return newFormatter(
		ssf.cfg,
		newTokenizer(ssf.cfg.TokenizerConfig),
		func(tok token, previousReservedWord token) token {
			if tok.typ == tokenTypeReservedTopLevel && tok.value == "SET" && previousReservedWord.value == "BY" {
				tok.typ = tokenTypeReserved
			}
			return tok
		},
	).format(query)
}
