#!/bin/bash
# Migration script: Remove old AST/lexer/parser, replace with EBNF pipeline
# Run from aiki project root

set -e

echo "=== Aiki EBNF Migration ==="
echo ""

# 1. Back up old packages (optional)
# mkdir -p .backup
# cp -r lang/ast lang/lexer lang/parser lang/token .backup/

# 2. Replace modified files
echo "Replacing lang/value/value.go..."
cp migration/lang/value/value.go lang/value/value.go

echo "Replacing lang/eval/eval_node.go..."
cp migration/lang/eval/eval_node.go lang/eval/eval_node.go

echo "Replacing cmd/fmt/ (stubbed)..."
rm -f cmd/fmt/builtin.go cmd/fmt/cmd.go cmd/fmt/files.go cmd/fmt/format.go
cp migration/cmd/fmt/builtin.go cmd/fmt/
cp migration/cmd/fmt/cmd.go cmd/fmt/

echo "Replacing cmd/lint/ (stubbed)..."
rm -f cmd/lint/builtin.go cmd/lint/cmd.go cmd/lint/files.go cmd/lint/rules.go
cp migration/cmd/lint/builtin.go cmd/lint/
cp migration/cmd/lint/cmd.go cmd/lint/

# 3. Delete old eval.go
echo "Deleting lang/eval/eval.go..."
rm -f lang/eval/eval.go

# 4. Delete old packages
echo "Deleting lang/ast/..."
rm -rf lang/ast

echo "Deleting lang/lexer/..."
rm -rf lang/lexer

echo "Deleting lang/parser/..."
rm -rf lang/parser

echo "Deleting lang/token/..."
rm -rf lang/token

# 5. Delete old tests that depend on removed packages
echo "Deleting tests that depend on old packages..."
rm -f tests/lexer_test.go
rm -f tests/parser_test.go
rm -f tests/fmt_test.go

# 6. Build to verify
echo ""
echo "Building..."
if go build ./cmd; then
    echo "✓ Build succeeded!"
else
    echo "✗ Build failed"
    exit 1
fi

# 7. Run tests
echo ""
echo "Running tests..."
if go test ./...; then
    echo "✓ Tests passed!"
else
    echo "✗ Some tests failed"
fi

echo ""
echo "=== Migration complete ==="
echo ""
echo "Stubbed commands (need reimplementation):"
echo "  - aiki fmt"
echo "  - aiki lint"
echo ""
echo "Deleted packages:"
echo "  - lang/ast"
echo "  - lang/lexer"
echo "  - lang/parser"
echo "  - lang/token"
echo "  - lang/eval/eval.go"
