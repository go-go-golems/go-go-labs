## Next steps
- [ ] properly count files when computing stats
- [ ] allow loading of filter rules from a .catter.yaml file

## Glazed converstion

- [ ] convert to glazed command
- [ ] print out stats using glazed
- [ ] make a custom verb for the stats
- [ ] turn filter into a utility package
- [ ] add glazed flag layer for filter options (to reuse in other commands)

## Ideas

### Stats

- [ ] Filter out binary files
- [ ] Verify gitignore to ignore .history for example

## Done

- [x] Add more succinct flags
- [x] fix bug when computing stats
- [x] Show token count per directory / file type
- [x] AST walker generic per file / per directory (when called with directory, given list of files in the directory, potentially also the results for each of the files in that directory first).
