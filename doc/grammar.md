# Aiki Grammar

## Tokens

```
KEYWORD     let if else while match return true false and or not
NUMBER      42  3.14  1/3
STRING      "hello"
RUNE        'a'
SYMBOL      :name
SHAPE       @name
NAME        identifier
COMMENT     # to end of line
```

## Operators

```
+  -  *  /
<  >  <=  >=
and  or  not
|>
.
=
```

## Delimiters

```
(  )  [  ]  {  }  ,  ;  ...
```

## Grammar

```ebnf
program     = { statement [ ";" ] }

statement   = let_stmt | assign_stmt | if_stmt | while_stmt
            | match_stmt | return_stmt | expr_stmt

let_stmt    = "let" NAME "=" expr
            | "let" SHAPE "[" [ field { "," field } ] "]"
field       = NAME | SHAPE

assign_stmt = NAME "=" expr

if_stmt     = "if" expr block [ "else" ( if_stmt | block ) ]

while_stmt  = "while" expr block

match_stmt  = "match" expr "{" { pattern block } "}"

pattern     = "_" | NAME | literal
            | "[" [ pattern { "," pattern } ] "]"
            | "[" SHAPE { "," pattern } "]"

return_stmt = "return" expr

expr_stmt   = expr

expr        = pipe_expr

pipe_expr   = infix_expr { "|>" call_expr }

infix_expr  = unary_expr { BINOP unary_expr }

unary_expr  = [ "not" | "-" ] postfix_expr

postfix_expr = primary { call | index | access }

call        = "(" [ expr { "," expr } ] ")"

index       = "[" expr "]"

access      = "." NAME

primary     = NUMBER | STRING | RUNE | SYMBOL | NAME
            | "true" | "false"
            | list_literal
            | func_literal
            | "(" expr ")"

list_literal = "[" [ expr { "," expr } ] "]"
             | "[" SHAPE { "," expr } "]"

func_literal = "(" params ")" block

params      = [ param_list ] [ rest_param ]
            | rest_param

param_list  = NAME { "," NAME }

rest_param  = "..." NAME

block       = "{" { statement [ ";" ] } "}"

literal     = NUMBER | STRING | RUNE | SYMBOL | "true" | "false"

BINOP       = "+" | "-" | "*" | "/"
            | "<" | ">" | "<=" | ">="
            | "and" | "or"
```

## Notes

- No operator precedence. Left-to-right evaluation. Use parens for grouping.
- Commas separate function arguments, list elements, parameters, and pattern elements.
- Semicolons optionally separate statements on the same line.
- `let` creates bindings. `=` mutates existing bindings.
- Every function path requires explicit `return`.
- Rest parameters (`...name`) collect remaining arguments into a list.
- Bracket indexing works on lists, strings, and call results: `list[i]`, `s[0]`, `f()[0]`.
- Dot access is field-only (NAME), not numeric.
- Comments use `#` to end of line. No block comments.
- Modules use `import()` and `export()` functions, not keywords.
