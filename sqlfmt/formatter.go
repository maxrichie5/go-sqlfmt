package sqlfmt

import (
	"fmt"
	"regexp"
	"strings"
)

var limitKeywordRegex = regexp.MustCompile(`(?i)^LIMIT$`)

// trimSpacesEnd removes trailing spaces and tabs from a string.
func trimSpacesEnd(str string) string {
	return strings.TrimRight(str, " \t")
}

// formatter formats SQL queries for better readability.
type formatter struct {
	cfg                  *Config
	indentation          *indentation
	inlineBlock          *inlineBlock
	params               *params
	tokenizer            *tokenizer                                        // Assume tokenizer is defined in your code
	tokenOverride        func(tok token, previousReservedWord token) token // Assume token is defined in your code
	previousReservedWord token
	tokens               []token
	index                int
}

// newFormatter creates a new formatter instance.
func newFormatter(cfg *Config, tokenizer *tokenizer, tokenOverride func(tok token, previousReservedWord token) token) *formatter {
	if cfg.ColorConfig == nil {
		cfg.ColorConfig = &ColorConfig{}
	}
	return &formatter{
		cfg:                  cfg,
		indentation:          newIndentation(cfg.Indent),
		inlineBlock:          newInlineBlock(),
		params:               newParams(cfg.Params),
		tokenizer:            tokenizer,
		tokenOverride:        tokenOverride,
		previousReservedWord: token{},
		tokens:               []token{},
		index:                0,
	}
}

// format formats whitespace in a SQL string to make it easier to read.
func (f *formatter) format(query string) string {
	f.tokens = f.tokenizer.tokenize(query)
	formattedQuery := f.getFormattedQueryFromTokens()
	return strings.TrimSpace(formattedQuery)
}

// getFormattedQueryFromTokens processes the tokens to create a formatted query.
func (f *formatter) getFormattedQueryFromTokens() string {
	// TODO: replace with string builder
	formattedQuery := ""

	for i, tok := range f.tokens {
		if tok.typ != tokenTypeWhitespace {
			fmt.Println(tok.typ, tok.value)
		}
		f.index = i

		if f.tokenOverride != nil {
			tok = f.tokenOverride(tok, f.previousReservedWord)
		}

		switch tok.typ {
		case tokenTypeWhitespace:
			// Ignore whitespace
		case tokenTypeLineComment:
			formattedQuery = f.formatLineComment(tok, formattedQuery)
		case tokenTypeBlockComment:
			formattedQuery = f.formatBlockComment(tok, formattedQuery)
		case tokenTypeReservedTopLevel:
			formattedQuery = f.formatTopLevelReservedWord(tok, formattedQuery)
			f.previousReservedWord = tok
		case tokenTypeReservedTopLevelNoIndent:
			formattedQuery = f.formatTopLevelReservedWordNoIndent(tok, formattedQuery)
			f.previousReservedWord = tok
		case tokenTypeReservedNewline:
			formattedQuery = f.formatNewlineReservedWord(tok, formattedQuery)
			f.previousReservedWord = tok
		case tokenTypeReserved:
			formattedQuery = f.formatWithSpaces(tok, formattedQuery)
			f.previousReservedWord = tok
		case tokenTypeOpenParen:
			formattedQuery = f.formatOpeningParentheses(tok, formattedQuery)
		case tokenTypeCloseParen:
			formattedQuery = f.formatClosingParentheses(tok, formattedQuery)
		case tokenTypeWord, tokenTypePlaceholder:
			if f.nextToken().typ == tokenTypePlaceholder {
				formattedQuery += tok.value
			} else if tok.typ == tokenTypePlaceholder {
				formattedQuery = f.formatPlaceholder(tok, formattedQuery)
			} else {
				formattedQuery = f.formatWithSpaces(tok, formattedQuery)
			}
		case tokenTypeString:
			formattedQuery = f.formatString(tok, formattedQuery)
		case tokenTypeNumber:
			formattedQuery = f.formatNumber(tok, formattedQuery)
		case tokenTypeBoolean:
			formattedQuery = f.formatBoolean(tok, formattedQuery)
		default:
			switch tok.value {
			case ",":
				formattedQuery = f.formatComma(tok, formattedQuery)
			case ":":
				formattedQuery = f.formatWithSpaceAfter(tok, formattedQuery)
			case ".":
				formattedQuery = f.formatWithoutSpaceAfter(tok, formattedQuery)
			case ";":
				formattedQuery = f.formatQuerySeparator(tok, formattedQuery)
			default:
				formattedQuery = f.formatWithSpaces(tok, formattedQuery)
			}
		}
	}
	return formattedQuery
}

func (f *formatter) formatLineComment(tok token, query string) string {
	value := tok.value
	value = addANSIFormats(f.cfg.ColorConfig.CommentFormatOptions, value)
	return f.addNewline(query + value)
}

func (f *formatter) formatBlockComment(tok token, query string) string {
	value := tok.value
	value = f.indentComment(value)
	value = addANSIFormats(f.cfg.ColorConfig.CommentFormatOptions, value)
	return f.addNewline(f.addNewline(query) + value)
}

func (f *formatter) indentComment(comment string) string {
	return regexp.MustCompile(`\n[ \t]*`).ReplaceAllString(comment, "\n"+f.indentation.getIndent()+" ")
}

func (f *formatter) formatTopLevelReservedWordNoIndent(tok token, query string) string {
	f.indentation.decreaseTopLevel()
	query = f.addNewline(query) + f.equalizeWhitespace(f.formatReservedWord(tok.value))
	return f.addNewline(query)
}

func (f *formatter) formatTopLevelReservedWord(tok token, query string) string {
	f.indentation.decreaseTopLevel()
	query = f.addNewline(query)

	f.indentation.increaseTopLevel()
	query += f.equalizeWhitespace(f.formatReservedWord(tok.value))

	return f.addNewline(query)
}

func (f *formatter) formatNewlineReservedWord(tok token, query string) string {
	return f.addNewline(query) + f.equalizeWhitespace(f.formatReservedWord(tok.value)) + " "
}

// equalizeWhitespace replaces any sequence of whitespace characters with a single space.
func (f *formatter) equalizeWhitespace(s string) string {
	return regexp.MustCompile(`\s+`).ReplaceAllString(s, " ")
}

func (f *formatter) formatOpeningParentheses(tok token, query string) string {
	preserveWhitespaceFor := map[tokenType]struct{}{
		tokenTypeWhitespace:  {},
		tokenTypeOpenParen:   {},
		tokenTypeLineComment: {},
	}

	if _, ok := preserveWhitespaceFor[f.previousToken().typ]; !ok {
		query = trimSpacesEnd(query)
	}

	value := tok.value
	if f.cfg.Uppercase {
		value = strings.ToUpper(value)
	}
	query += value

	f.inlineBlock.beginIfPossible(f.tokens, f.index)
	if !f.inlineBlock.isActive() {
		f.indentation.increaseBlockLevel()
		query = f.addNewline(query)
	}
	return query
}

// formatClosingParentheses ends an inline block if one is active, or decreases the
// block level, then adds the closing paren.
func (f *formatter) formatClosingParentheses(tok token, query string) string {
	if f.cfg.Uppercase {
		tok.value = strings.ToUpper(tok.value)
	}

	if f.inlineBlock.isActive() {
		f.inlineBlock.end()
		return f.formatWithSpaceAfter(tok, query)
	} else {
		f.indentation.decreaseBlockLevel()
		return f.formatWithSpaces(tok, f.addNewline(query))
	}
}

// formatPlaceholder formats a placeholder by replacing it with a param value
// from the cfg params and adds a space after.
func (f *formatter) formatPlaceholder(tok token, query string) string {
	return query + f.params.get(tok.key, tok.value) + " "
}

// formatComma adds the comma to the query and adds a space. If an inline block
// is not active, it will add a new line too.
func (f *formatter) formatComma(tok token, query string) string {
	query = trimSpacesEnd(query) + tok.value + " "

	if f.inlineBlock.isActive() {
		return query
	} else if limitKeywordRegex.MatchString(f.previousReservedWord.value) {
		// avoids creating new lines after LIMIT keyword so that two limit items appear on one line for nicer formatting
		return query
	} else {
		return f.addNewline(query)
	}
}

// formatWithSpaceAfter returns the query with spaces trimmed off the end,
// the token value, and a space (" ") at the end ("query value ")
func (f *formatter) formatWithSpaceAfter(tok token, query string) string {
	return trimSpacesEnd(query) + tok.value + " "
}

// formatWithoutSpaceAfter returns the query with spaces trimmed off the end and
// the token value ("query value")
func (f *formatter) formatWithoutSpaceAfter(tok token, query string) string {
	return trimSpacesEnd(query) + tok.value
}

// TODO: this can probably be replaced with formatWithSpaceAfter
func (f *formatter) formatWithSpaces(tok token, query string) string {
	value := tok.value
	if tok.typ == tokenTypeReserved {
		value = f.formatReservedWord(tok.value)
	}

	next := f.nextToken()
	if tok.typ == tokenTypeWord && !next.empty() && next.value == "(" {
		value = addANSIFormats(f.cfg.ColorConfig.FunctionCallFormatOptions, value)
	}

	return query + value + " "
}

// formatReservedWord makes sure the reserved word is formatted according to the Config.
func (f *formatter) formatReservedWord(value string) string {
	if f.cfg.Uppercase {
		value = strings.ToUpper(value)
	}
	value = addANSIFormats(f.cfg.ColorConfig.ReservedWordFormatOptions, value)
	return value
}

func (f *formatter) formatQuerySeparator(tok token, query string) string {
	f.indentation.resetIndentation()
	return trimSpacesEnd(query) + tok.value + strings.Repeat("\n", f.cfg.LinesBetweenQueries)
}

func (f *formatter) formatString(tok token, query string) string {
	value := tok.value
	value = addANSIFormats(f.cfg.ColorConfig.StringFormatOptions, value)
	return query + value + " "
}

func (f *formatter) formatNumber(tok token, query string) string {
	value := tok.value
	value = addANSIFormats(f.cfg.ColorConfig.NumberFormatOptions, value)
	return query + value + " "
}

func (f *formatter) formatBoolean(tok token, query string) string {
	value := tok.value
	value = addANSIFormats(f.cfg.ColorConfig.BooleanFormatOptions, value)
	return query + value + " "
}

// addNewline trims spaces from the end of query, adds a new line character if
// one does not already exist at the end, and adds the indentation to the new
// line.
func (f *formatter) addNewline(query string) string {
	query = trimSpacesEnd(query)
	if !strings.HasSuffix(query, "\n") {
		query += "\n"
	}
	return query + f.indentation.getIndent()
}

// previousToken peeks at the previous token in the formatters list of tokens with
// the given offset. If no offset is provided, a default of 1 is used.
func (f *formatter) previousToken(offset ...int) token {
	o := 1
	if len(offset) > 0 {
		o = offset[0]
	}
	if f.index-o < 0 {
		return token{} // return an empty token struct
	}
	return f.tokens[f.index-o]
}

// nextToken peeks at the next token in the formatters list of tokens with
// the given offset. If no offset is provided, a default of 1 is used. If
// there is no next token, it returns an empty token.
func (f *formatter) nextToken(offset ...int) token {
	o := 1
	if len(offset) > 0 {
		o = offset[0]
	}
	if f.index+o >= len(f.tokens) {
		return token{}
	}
	return f.tokens[f.index+o]
}
