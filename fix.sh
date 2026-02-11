#!/bin/bash
set -e

# Update package declarations
sed -i 's/^package hal$/package core/' hal/core/*.go
# hal/canvas stays package canvas
# lang/* packages keep their names
# cmd/* packages keep their names

# Update imports in all Go files
find . -name "*.go" -type f | xargs sed -i 's|"aiki/hal"|"aiki/hal/core"|g'
find . -name "*.go" -type f | xargs sed -i 's|"aiki/canvas"|"aiki/hal/canvas"|g'
find . -name "*.go" -type f | xargs sed -i 's|"aiki/ast"|"aiki/lang/ast"|g'
find . -name "*.go" -type f | xargs sed -i 's|"aiki/lexer"|"aiki/lang/lexer"|g'
find . -name "*.go" -type f | xargs sed -i 's|"aiki/parser"|"aiki/lang/parser"|g'
find . -name "*.go" -type f | xargs sed -i 's|"aiki/token"|"aiki/lang/token"|g'
find . -name "*.go" -type f | xargs sed -i 's|"aiki/eval"|"aiki/lang/eval"|g'
find . -name "*.go" -type f | xargs sed -i 's|"aiki/value"|"aiki/lang/value"|g'
find . -name "*.go" -type f | xargs sed -i 's|"aiki/subcommands/fmt"|"aiki/cmd/fmt"|g'
find . -name "*.go" -type f | xargs sed -i 's|"aiki/subcommands/lint"|"aiki/cmd/lint"|g'
find . -name "*.go" -type f | xargs sed -i 's|"aiki/repl"|"aiki/cmd/repl"|g'

# Fix hal.X references -> core.X
find . -name "*.go" -type f | xargs sed -i 's|hal\.HAL|core.HAL|g'
find . -name "*.go" -type f | xargs sed -i 's|hal\.CloseAllCanvases|core.CloseAllCanvases|g'
find . -name "*.go" -type f | xargs sed -i 's|hal\.Reset\b|core.Reset|g'
find . -name "*.go" -type f | xargs sed -i 's|hal\.ResetSignal|core.ResetSignal|g'
find . -name "*.go" -type f | xargs sed -i 's|hal\.LastPrintEndedWithNewline|core.LastPrintEndedWithNewline|g'
find . -name "*.go" -type f | xargs sed -i 's|hal\.ResetLastPrint|core.ResetLastPrint|g'
find . -name "*.go" -type f | xargs sed -i 's|hal\.DefaultScheduler|core.DefaultScheduler|g'
find . -name "*.go" -type f | xargs sed -i 's|hal\.NativeBoolToBoolean|core.NativeBoolToBoolean|g'

echo "Done. Run 'go build' to check for errors."
