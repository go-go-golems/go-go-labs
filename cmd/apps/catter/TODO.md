- [x] Add more succinct flags

## Ideas

### Stats

- [x] Show token count per directory / file type
- [x] AST walker generic per file / per directory (when called with directory, given list of files in the directory, potentially also the results for each of the files in that directory first).

- [ ] fix bug when computing stats

```
❯ go run ./cmd/apps/catter --stats overview
Error computing stats: error creating filewalker: either fs.FS must be set or paths must not be empty
```

- [ ] Filter out binary files
- [ ] Verify gitignore to ignore .history for example
