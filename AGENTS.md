# Command Reference

## Build and Test Commands

To verify the code compiles and the tests pass, run:

```bash
go run ./cmd/cdktf-provider-gen/ -c ./example.yml -cdktf-version 0.20.7 
```

You can use `-keep` flag to keep the generated code for furthur inspection:

```bash
SRC_LOG_LEVEL=debug go run ./cmd/cdktf-provider-gen/ -c ./example.yml -cdktf-version 0.20.7 -keep
```

You can find the directory from logs. Remember to delete the directory after inspection to save disk space.

# Code Standards and Conventions

- Use `github.com/sourcegraph/run` for running external commands.
- Use `github.com/sourcegraph/log` for logging.
- Use `github.com/pkg/errors` for error handling.
