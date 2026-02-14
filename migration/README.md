# Aiki EBNF Migration

This tarball contains files to migrate from the old hand-written lexer/parser/AST to the EBNF-driven pipeline.

## What's Changed

### Modified Files
- `lang/value/value.go` - Removed `Body *ast.BlockStatement` field, removed ast import
- `lang/eval/eval_node.go` - Added `HAL`, `isError`, `isTruthy` helpers; removed `fn.Body` fallback
- `cmd/fmt/` - Stubbed (returns "not implemented")
- `cmd/lint/` - Stubbed (returns "not implemented")

### Deleted Files
- `lang/eval/eval.go` - Old evaluator
- `lang/ast/` - Old AST package
- `lang/lexer/` - Old lexer package
- `lang/parser/` - Old parser package
- `lang/token/` - Old token package
- `tests/lexer_test.go` - Tests old lexer
- `tests/parser_test.go` - Tests old parser
- `tests/fmt_test.go` - Tests old formatter

## How to Apply

### Option 1: Run the script
```bash
tar xzf migration.tar.gz
cd migration
./migrate.sh
```

### Option 2: Manual
```bash
tar xzf migration.tar.gz

# Replace files
cp migration/lang/value/value.go lang/value/
cp migration/lang/eval/eval_node.go lang/eval/
rm -f cmd/fmt/*.go && cp migration/cmd/fmt/*.go cmd/fmt/
rm -f cmd/lint/*.go && cp migration/cmd/lint/*.go cmd/lint/

# Delete old packages
rm -f lang/eval/eval.go
rm -rf lang/ast lang/lexer lang/parser lang/token
rm -f tests/lexer_test.go tests/parser_test.go tests/fmt_test.go

# Build
go build ./cmd
go test ./...
```

## After Migration

The following commands are stubbed and return errors:
- `aiki fmt <file>` - needs reimplementation with EBNF AST
- `aiki lint <file>` - needs reimplementation with EBNF AST

Everything else should work as before, using the new EBNF pipeline.
