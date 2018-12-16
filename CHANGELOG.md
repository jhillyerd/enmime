Change Log
==========

All notable changes to this project will be documented in this file.
This project adheres to [Semantic Versioning](http://semver.org/).

## [0.5.0] - 2018-12-15

### Added
- Use github.com/pkg/errors to decorate errors with stack traces (thanks to
  dcomier.)
- Several improvements to Content-Type header decoding (thanks to dcormier.)
- File modification date to encode/decode (thanks to dann7387.)
- Handle non-delimited address lists (thanks to requaos.)
- RFC-2047 attribute name deocding (thanks to requaos.)

### Fixed
- Only detect charset on `text/*` parts (thanks to dcormier.)
- Stop adding extra newline during encode (thanks to dann7387.)
- Math bug in selecting QP or base64 encoding (thanks to dann7387.)

## [0.4.0] - 2018-11-21

### Added
- Override declared character set if another is detected with high confidence
  (thanks to nerdlich.)
- Handle unquoted specials in media type parameters (thanks to requaos.)
- Handle barren Content-Type headers (thanks to dcormier.)
- Better handle malformed media type parameters (thanks to dcormier.)

### Changed
- Use iso-8859-1 character map when implicitly declared (thanks to requaos.)
- Treat "inline" disposition as message content, not attachment unless it is
  accompanied by parameters (e.g. a filename, thanks to requaos.)

## [0.3.0] - 2018-11-01

### Added
- CLI utils now output inlines and other parts in addition to attachments.
- Clone() method to Envelope and Part (thanks to nerdlich.)
- GetHeaderKeys() method to Envelope (thanks to allenluce.)
- GetHeaderValues() plus a suite of setters for Envelope (thanks to nerdlich.)

### Changed
- Use value instead of pointer receivers and return types on MailBuilder
  methods.  Cleaner API, but may break some users.
- `enmime.Error` now conforms to the Go error interface, its `String()` method
  is now deprecated.
- `NewPart()` constructor no longer takes a parent parameter.
- Part.Errors now holds pointers, matching Envelope.Errors.

### Fixed
- Content is now populated for binary-only mails root part (thank to ostcar.)

### Removed
- Part no longer implements `io.Reader`, content is stored as a byte slice in
  `Part.Content` instead.


## [0.2.1] - 2018-10-20

### Added
- Go modules support for reproducible builds.


## [0.2.0] - 2018-02-24

### Changed
- Encoded filenames now have unicode accents stripped instead of escaped, making
  them more readable.
- Part.ContentID
  - is now properly encoded into the headers when using the builder.
  - is now populated from headers when decoding messages.
- Update go doc, add info about headers and errors.

### Fixed
- Part.Read() and Part.Utf8Reader, they are deprecated but should continue to
  function until 1.0.0.


## 0.1.0 - 2018-02-10

### Added
- Initial implementation of MIME encoding, using `enmime.MailBuilder`

[Unreleased]: https://github.com/jhillyerd/enmime/compare/master...develop
[0.5.0]:      https://github.com/jhillyerd/enmime/compare/v0.4.0...v0.5.0
[0.4.0]:      https://github.com/jhillyerd/enmime/compare/v0.3.0...v0.4.0
[0.3.0]:      https://github.com/jhillyerd/enmime/compare/v0.2.1...v0.3.0
[0.2.1]:      https://github.com/jhillyerd/enmime/compare/v0.2.0...v0.2.1
[0.2.0]:      https://github.com/jhillyerd/enmime/compare/v0.1.0...v0.2.0


## Release Checklist

1.  Create release branch: `git flow release start 1.x.0`
2.  Update CHANGELOG.md:
    - Ensure *Unreleased* section is up to date.
    - Rename *Unreleased* section to release name and date.
    - Add new GitHub `/compare` link.
3.  Run tests
4.  Commit changes and merge release: `git flow release finish`

See http://keepachangelog.com/ for additional instructions on how to update this
file.
