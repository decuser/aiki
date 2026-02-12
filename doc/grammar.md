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

**Removed (provisional):**
- `==` → use `equal(a, b)`
- `!=` → use `not(equal(a, b))`
- `%` → use `modulo(a, b)`
- `!` → use `not`

## Delimiters

```
(  )  [  ]  {  }  ,
```

## Grammar

```ebnf
program     = { statement }

statement   = let_stmt | assign_stmt | if_stmt | while_stmt
            | match_stmt | return_stmt | expr_stmt

let_stmt    = "let" NAME "=" expr
            | "let" SHAPE "[" [ field { "," field } ] "]"
field       = NAME | SHAPE

assign_stmt = NAME "=" expr

if_stmt     = "if" expr block [ "else" block ]

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

unary_expr  = [ "not" | "-" ] call_expr

call_expr   = access_expr { "(" [ expr { "," expr } ] ")" }

access_expr = primary { "." ( NAME | NUMBER ) }

primary     = NUMBER | STRING | RUNE | SYMBOL | NAME
            | "true" | "false"
            | "[" [ expr { "," expr } ] "]"
            | "[" SHAPE { "," expr } "]"
            | "(" [ NAME { "," NAME } ] ")" block
            | "(" expr ")"

block       = "{" { statement } "}"

literal     = NUMBER | STRING | RUNE | SYMBOL | "true" | "false"

BINOP       = "+" | "-" | "*" | "/"
            | "<" | ">" | "<=" | ">="
            | "and" | "or"
```

## Notes

- No operator precedence. Left-to-right evaluation. Use parens for grouping.
- Commas separate function arguments, list elements, parameters, and pattern elements.
- `let` creates bindings. `=` mutates existing bindings.
- Every function path requires explicit `return`.
- Modules use `import()` and `export()` functions, not keywords.
