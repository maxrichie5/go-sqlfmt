package sqlfmt

import (
	"fmt"
	"github.com/stretchr/testify/require"
	"strings"
	"testing"
)

func TestN1QLFormatter_Format(t *testing.T) {
	tests := []struct {
		name  string
		query string
		exp   string
		cfg   Config
	}{
		{
			name:  "formats SELECT query with element selection expression",
			query: "SELECT order_lines[0].productId FROM orders;",
			exp: Dedent(`
              SELECT
                order_lines[0].productId
              FROM
                orders;
            `),
		},
		{
			name:  "formats SELECT query with primary key querying",
			query: "SELECT fname, email FROM tutorial USE KEYS ['dave', 'ian'];",
			exp: Dedent(`
              SELECT
                fname,
                email
              FROM
                tutorial
              USE KEYS
                ['dave', 'ian'];
            `),
		},
		{
			name:  "formats INSERT with {} object literal",
			query: "INSERT INTO heroes (KEY, VALUE) VALUES ('123', {'id':1,'type':'Tarzan'});",
			exp: Dedent(`
              INSERT INTO
                heroes (KEY, VALUE)
              VALUES
                ('123', {'id': 1, 'type': 'Tarzan'});
            `),
		},
		{
			name: "formats INSERT with large object and array literals",
			query: `
              INSERT INTO heroes (KEY, VALUE) VALUES ('123', {'id': 1, 'type': 'Tarzan',
              'array': [123456789, 123456789, 123456789, 123456789, 123456789], 'hello': 'world'});
            `,
			exp: Dedent(`
              INSERT INTO
                heroes (KEY, VALUE)
              VALUES
                (
                  '123',
                  {
                    'id': 1,
                    'type': 'Tarzan',
                    'array': [
                      123456789,
                      123456789,
                      123456789,
                      123456789,
                      123456789
                    ],
                    'hello': 'world'
                  }
                );
            `),
		},
		{
			name:  "formats SELECT query with UNNEST top level reserved word",
			query: "SELECT * FROM tutorial UNNEST tutorial.children c;",
			exp: Dedent(`
              SELECT
                *
              FROM
                tutorial
              UNNEST
                tutorial.children c;
            `),
		},
		{
			name: "formats SELECT query with NEST and USE KEYS",
			query: `
              SELECT * FROM usr
              USE KEYS 'Elinor_33313792' NEST orders_with_users orders
              ON KEYS ARRAY s.order_id FOR s IN usr.shipped_order_history END;
            `,
			exp: Dedent(`
              SELECT
                *
              FROM
                usr
              USE KEYS
                'Elinor_33313792'
              NEST
                orders_with_users orders ON KEYS ARRAY s.order_id FOR s IN usr.shipped_order_history END;
            `),
		},
		{
			name:  "formats explained DELETE query with USE KEYS and RETURNING",
			query: "EXPLAIN DELETE FROM tutorial t USE KEYS 'baldwin' RETURNING t",
			exp: Dedent(`
              EXPLAIN DELETE FROM
                tutorial t
              USE KEYS
                'baldwin' RETURNING t
            `),
		},
		{
			name:  "formats UPDATE query with USE KEYS and RETURNING",
			query: "UPDATE tutorial USE KEYS 'baldwin' SET type = 'actor' RETURNING tutorial.type",
			exp: Dedent(`
              UPDATE
                tutorial
              USE KEYS
                'baldwin'
              SET
                type = 'actor' RETURNING tutorial.type
            `),
		},
		{
			name:  "recognizes $variables",
			query: "SELECT $variable, $'var name', $\"var name\", $`var name`;",
			exp: Dedent(`
              SELECT
                $variable,
                $'var name',
                $"var name",
                ` + "$`var name`;" + `
            `),
		},
		{
			name:  "replaces $variables with param values",
			query: "SELECT $variable, $'var name', $\"var name\", $`var name`;",
			exp: Dedent(`
              SELECT
                "variable value",
                'var value',
                'var value',
                'var value';
            `),
			cfg: Config{
				Params: NewMapParams(map[string]string{
					"variable": `"variable value"`,
					"var name": "'var value'",
				}),
			},
		},
		{
			name:  "replaces $ numbered placeholders with param values",
			query: "SELECT $1, $2, $0;",
			exp: Dedent(`
              SELECT
                second,
                third,
                first;
            `),
			cfg: Config{
				Params: NewListParams([]string{
					"first",
					"second",
					"third",
				}),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var result string
			if !tt.cfg.Empty() {
				if tt.cfg.Indent == "" {
					tt.cfg.Indent = DefaultIndent
				}
				result = NewN1QLFormatter(&tt.cfg).Format(tt.query)
			} else {
				result = NewN1QLFormatter(NewDefaultConfig()).Format(tt.query)
			}

			exp := strings.TrimRight(tt.exp, "\n\t ")
			exp = strings.TrimLeft(exp, "\n")
			exp = strings.ReplaceAll(exp, "\t", DefaultIndent)

			if result != exp {
				fmt.Println("=== QUERY ===")
				fmt.Println(tt.query)
				fmt.Println()

				fmt.Println("=== EXP ===")
				fmt.Println(exp)
				fmt.Println()

				fmt.Println("=== RESULT ===")
				fmt.Println(result)
				fmt.Println()
			}
			require.Equal(t, exp, result)
		})
	}
}
