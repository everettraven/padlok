# Third Party Packages

This internal package is meant to carry forks of third-party libraries that we need make
modifications to in order to support our desired functionality.

To prevent significant deviations and simplify the process of rebasing these forked
libraries, all commits must follow the following guidelines:

- Commits that are unique to the `padlok` project and will need to be carried ~forever must be prefixed with `UPSTREAM: <carry>:` (`<carry>` is the literal string and should not be substituted)
- Commits that have an not-yet-merged Pull Request but we need to pick over before it is merged must be prefixed with `UPSTREAM: {PR number}:`, substituting `{PR number}` with the upstream PR number
- Commits that are specific to the `padlok` project and are temporary and can be dropped in a future rebase (like generated files) must be prefixed with `UPSTREAM: <drop>` (`<drop>` is the literal string and should not be substituted)

All forked packages _MUST_ contain a `LICENSE` file that matches the forked library license _AND_ a `NOTICE.md` file that documents that the files have been modified.

Changes to forked files should be kept to what is minimally required to achieve the goals of the changes.
