# BUG_REPRO

The following failures were observed while validating the initial project state.
Each section records what failed, how to reproduce it, and the complete command output.
They are preserved intentionally; only failing build gates are omitted from the generated Dockerfile.

## Failure 1: Go test (.)

- Observed problem: `Go test (.)` failed in the initial project state.
- Working directory: `.`
- Command: `cd /app && GOTOOLCHAIN=local GOPROXY=off GOSUMDB=off go test -count=1 ./...`
- Exit status: `1`

```text
?   	example.com/animalcage/cmd/cage-service	[no test files]
ok  	example.com/animalcage/internal/backup	0.001s
--- FAIL: Test636BusinessRegression (0.00s)
    regression_test.go:55: second roster confirmation = "mouse-a", want latest name
FAIL
FAIL	example.com/animalcage/internal/flow004	0.001s
?   	example.com/animalcage/internal/model	[no test files]
ok  	example.com/animalcage/internal/httpapi	0.002s
ok  	example.com/animalcage/internal/notification	0.001s
ok  	example.com/animalcage/internal/planning	0.001s
ok  	example.com/animalcage/internal/registry	0.001s
ok  	example.com/animalcage/internal/report	0.001s
ok  	example.com/animalcage/internal/store	0.018s
ok  	example.com/animalcage/internal/validation	0.001s
ok  	example.com/animalcage/internal/workflow	0.001s
FAIL
```

## Architecture reproduction

### linux/amd64
- Go toolchain version: exit `0`
- Node.js version: exit `0`
- Go build (.): exit `0`
- Go test (.): exit `1`
- Go run smoke (cmd/cage-service): exit `0`
- Frontend build (web): exit `0`
### linux/arm64
- Go toolchain version: exit `0`
- Node.js version: exit `0`
- Go build (.): exit `0`
- Go test (.): exit `1`
- Go run smoke (cmd/cage-service): exit `0`
- Frontend build (web): exit `0`
