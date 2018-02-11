Change Log
==========

All notable changes to this project will be documented in this file.
This project adheres to [Semantic Versioning](http://semver.org/).

## 0.1 - 2018-02-10

### Added
- Initial implementation of MIME encoding, using `enmime.MailBuilder`

[Unreleased]: https://github.com/jhillyerd/enmime/compare/master...develop


## Release Checklist

1.  Create release branch: `git flow release start 1.x.0`
2.  Update CHANGELOG.md:
    - Ensure *Unreleased* section is up to date
    - Rename *Unreleased* section to release name and date.
    - Add new GitHub `/compare` link
3.  Run tests
4.  Commit changes and merge release: `git flow release finish`

See http://keepachangelog.com/ for additional instructions on how to update this
file.
