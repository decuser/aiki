#!/bin/bash
set -e  # Abort immediately if any command fails
set -x  # Print each command to the terminal before running it

# 1. Build and verify
make build
make fmt
make lint
make doclint

# 2. Run unit tests
make test

# 3. Run graphical test (opens window for 1s then closes)
./aiki examples/canvas.ai

# 4. Run batch examples
./aiki examples/test_hash.ai
./aiki examples/newton-imp.ai
./aiki examples/functional.ai

# 5. Run interactive example (pipe input "42" to prevent hanging)
echo "42" | ./aiki examples/game.ai

# 6. Run CLI one-liners
./aiki -e 'if 1+1 == 2 { print("Math OK\n") }'
./aiki -e 'import "examples/fib.ai"; if fib(10) == 55 { print("Import/Func OK\n") }'

# If we get here, everything passed
set +x
echo "Validation Complete."
