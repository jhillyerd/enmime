How to Contribute
=================

Enmime high encourages third-party patches. There is a great deal of MIME email
out there, so it's likely you will encounter a scenario we haven't.


## Getting Started

If you anticipate your issue requiring a large patch, please submit a GitHub
issue describing the problem or feature first. Please attach an email that
illustrates the scenario you are trying to improve if possible. You are
encouraged to outline the process you would like to use to resolve the issue. I
will attempt to provide prompt validation and/or guidance on your suggested
approach.


## Making Changes

- Create a topic branch from where you want to base your work.
  - This is usually the `develop` branch.
  - Only target the `master` branch if the issue is already resolved in
    `develop`.
- Make commits of logical units.
- Add unit tests to exercise your changes.
- Run the updated code through `go fmt` and `go vet`.
- Ensure the code builds and tests with the following commands:
  - `go clean ./...`
  - `go build ./...`
  - `go test ./...`


## Thanks

Thank you for considering contributing to enmime!
