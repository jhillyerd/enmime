How to Contribute
=================

Enmime highly encourages third-party patches. There is a great deal of MIME
encoded email out there, so it's likely you will encounter a scenario we
haven't.

**tl;dr:** File pull requests against the `develop` branch, not `master`!


## Getting Started

If you anticipate your issue requiring a large patch, please first submit a
GitHub issue describing the problem or feature. Attach an email that illustrates
the scenario you are trying to improve if possible. You are also encouraged to
outline the process you would like to use to resolve the issue. I will attempt
to provide validation and/or guidance on your suggested approach.


## Making Changes

Enmime uses [git-flow] with default options.  If you have git-flow installed,
you can run `git flow feature start <topic branch name>`.

Without git-flow, create a topic branch from where you want to base your work:
  - This is usually the `develop` branch, example command:
    `git checkout origin/develop -b <topic branch name>`
  - Only target the `master` branch if the issue is already resolved in
    `develop`.

Once you are on your topic branch:

1. Make commits of logical units.
2. Add unit tests to exercise your changes.
3. Scrub personally identifying information from test case emails, and
   keep attachments short.
4. Run the updated code through `go fmt` and `go vet`.
5. Ensure the code builds and tests with the following commands:
  - `go clean ./...`
  - `go build ./...`
  - `go test ./...`


## Thanks

Thank you for contributing to enmime!

[git-flow]: https://github.com/nvie/gitflow
