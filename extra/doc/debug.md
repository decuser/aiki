# aiki debug

The `debug` subcommand provides visibility into the lexer, parser, and evaluator stages.

## Usage

```
aiki debug [-stage lex|parse|eval|all] [-trace] [-prelude] <filename>
```

## Flags

| Flag | Default | Description |
|------|---------|-------------|
| `-stage` | `all` | Which stage to show: `lex`, `parse`, `eval`, or `all` |
| `-trace` | off | Enable observer output (detailed step-by-step) |
| `-prelude` | off | Include prelude in trace (requires `-trace`) |

## Examples


```
vi test.ai 
let x = 1 + 2; println(x)

# run the traces
./aiki debug -stage lex test.ai      # show tokens
./aiki debug -stage parse test.ai    # show AST
./aiki debug -stage eval test.ai     # run and show result
./aiki debug test.ai                 # all (default)
./aiki debug -trace test.ai          # with observer output
./aiki debug -trace -prelude test.ai   # include prelude
```

Given a test file:

```
let x = 1 + 2; println(x)
```

### Show tokens

```
$ ./aiki debug -stage lex test.ai
==== Tokens ====
1:1   KEYWORD    "let"
1:5   NAME       "x"
1:7   OPERATOR   "="
1:9   NUMBER     "1"
1:11  OPERATOR   "+"
1:13  NUMBER     "2"
1:14  DELIMITER  ";"
1:16  NAME       "println"
1:23  DELIMITER  "("
1:24  NAME       "x"
1:25  DELIMITER  ")"
```

### Show AST

```
$ ./aiki debug -stage parse test.ai
==== AST ====
program 1:1
  statement 1:1
    let_stmt 1:1
      TERMINAL 1:1 "let"
      NAME 1:5 "x"
      TERMINAL 1:7 "="
      expr 1:9
        pipe_expr 1:9
          infix_expr 1:9
            unary_expr 1:9
              postfix_expr 1:9
                primary 1:9
                  NUMBER 1:9 "1"
            BINOP 1:11
              TERMINAL 1:11 "+"
            unary_expr 1:13
              postfix_expr 1:13
                primary 1:13
                  NUMBER 1:13 "2"
  ...
```

### Run and show result

```
$ ./aiki debug -stage eval test.ai
==== Eval ====
3
=> []
```

### All stages (default)

```
$ ./aiki debug test.ai
==== Eval ====
3
=> []
```

### Trace observer output

```
$ ./aiki debug -trace test.ai
lex: KEYWORD "let" at 1:1
lex: NAME "x" at 1:5
...
parse: program at 1:1
parse: statement at 1:1
parse: let_stmt at 1:1
...
==== Eval ====
eval: program ->  (scope 0) at 1:1
eval: statement ->  (scope 0) at 1:1
eval: let_stmt ->  (scope 0) at 1:1
...
3
=> []
```

By default, `-trace` filters to the user file only. Calls into prelude functions are not shown.

### Include prelude in trace

```
$ ./aiki debug -trace -prelude test.ai
```

This shows all observer events including prelude loading and function calls into prelude-defined functions like `println` and `to_str`.
