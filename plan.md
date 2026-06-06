# Plan: Allow Forcing Content-Transfer-Encoding on a Part

**Issue:** [#395](https://github.com/jhillyerd/enmime/issues/395)  
**Approach:** Option 2 — Add a `ContentTransferEncoding` field to `Part`

## Problem

`setupMIMEHeaders()` in `encode.go` defaults to `teBase64` for all non-text content types, bypassing `selectTransferEncoding()` entirely. This means ASCII-safe content like PGP armor (`application/pgp-encrypted`, `application/pgp-signature`) is unnecessarily base64-encoded, with no way for the caller to override.

## Changes

### 1. Add field to `Part` struct — `part.go`

```go
type Part struct {
    // ... existing fields ...

    // ContentTransferEncoding forces the Content-Transfer-Encoding header to the specified
    // value when encoding this part. Valid values are "7bit", "8bit", "base64", and
    // "quoted-printable". When empty, the encoding is selected automatically.
    // Unrecognised values fall back to automatic detection.
    ContentTransferEncoding string
}
```

This is consistent with the existing pattern of exported string fields like `Charset`, `Boundary`, `ContentID`, etc.

### 2. Respect the field in `setupMIMEHeaders()` — `encode.go`

Modify the content transfer encoding selection block (lines ~108–135) to check `p.ContentTransferEncoding` **before** the automatic detection logic:

```go
cte := te7Bit
if len(p.Content) > 0 {
    // Check for explicit override first.
    if forced := p.resolveForcedCTE(); forced != teRaw {
        cte = forced
    } else if strings.Index(strings.ToLower(p.ContentType), "message/") == 0 {
        cte = te8Bit
    } else {
        cte = teBase64
        if p.TextContent() && p.ContentReader == nil {
            cte = p.selectTransferEncoding(p.Content, false)
            if p.Charset == "" {
                p.Charset = utf8
            }
        }
    }
    // ... existing header.Set switch ...
}
```

The override takes precedence over both `TextContent()` gating and `Encoder.ForceQuotedPrintableCte`, since it is the most specific (per-part) directive. `ContentReader` parts can also be overridden (currently they are hardcoded to base64 via `encodeContentFromReader`).

### 3. Add helper to resolve forced CTE — `encode.go`

```go
// resolveForcedCTE maps the ContentTransferEncoding field to an internal transferEncoding
// value. Returns teRaw (sentinel for "not set / unrecognised") when the field is empty or
// contains an unrecognised value.
func (p *Part) resolveForcedCTE() transferEncoding {
    switch strings.ToLower(p.ContentTransferEncoding) {
    case cte7Bit:
        return te7Bit
    case cte8Bit:
        return te8Bit
    case cteQuotedPrintable:
        return teQuoted
    case cteBase64:
        return teBase64
    default:
        return teRaw // sentinel: no override
    }
}
```

Using `teRaw` as the "not set" sentinel avoids introducing a new sentinel type, since `teRaw` is not a valid CTE value and is already used elsewhere for a similar "passthrough" meaning.

### 4. Handle `encodeContent` for non-base64 with `ContentReader` — `encode.go`

Currently `encodeContentFromReader` always base64-encodes. When a forced CTE of `te7Bit`, `te8Bit`, or `teQuoted` is used with a `ContentReader`, we need to either:

- **Option A (recommended):** Read all content from the reader into `p.Content` and use the standard `encodeContent` path. This is simple and correct. The caller forcing a CTE on a streaming part is an unusual case.
- **Option B:** Add streaming encode paths for QP and passthrough. Adds complexity for minimal benefit.

For the initial implementation, go with Option A: if `ContentReader != nil` and the forced CTE is not base64, slurp the reader into `p.Content` in `Encode()` before calling `setupMIMEHeaders()`, and set `ContentReader = nil`.

### 5. Add `WithContentTransferEncoding` on `MailBuilder` — `builder.go`

So users of the builder API can set the encoding on attachment/inline parts. Add a method to `Part` and a corresponding `MailBuilder` method to set it when adding parts.

On `Part`:

```go
// SetContentTransferEncoding sets the ContentTransferEncoding field.
func (p *Part) SetContentTransferEncoding(cte string) *Part {
    p.ContentTransferEncoding = cte
    return p
}
```

This mirrors the existing `WithEncoder()` builder-style pattern on `Part`.

No changes to `MailBuilder` itself in the initial PR — callers can set it on the `*Part` before passing it to `AddAttachmentPart()` / `AddInlinePart()`. A `MailBuilder` convenience method can be added later if demand exists.

## Tests — `encode_test.go`

Add the following test cases:

1. **Non-text part with 7bit-safe content, CTE forced to `"7bit"`**  
   Verify no base64 encoding is applied and the header reads `Content-Transfer-Encoding: 7bit`.

2. **Non-text part with CTE forced to `"quoted-printable"`**  
   Verify QP encoding is applied.

3. **Non-text part with CTE forced to `"base64"`**  
   Verify base64 is applied (same as default behavior).

4. **Non-text part with CTE forced to `"8bit"`**  
   Verify 8bit header is set and content is written verbatim.

5. **Non-text part with empty `ContentTransferEncoding`**  
   Verify existing default behavior (base64) is preserved.

6. **Non-text part with unrecognised `ContentTransferEncoding` value**  
   Verify fallback to automatic detection (base64 for non-text).

7. **Case-insensitive values**  
   Verify `"7Bit"`, `"BASE64"`, etc. all work.

8. **Text part with CTE forced to `"base64"`**  
   Verify that the override takes precedence over auto-detection (which would normally pick 7bit or QP for ASCII content).

9. **Part with `ContentReader` and forced `"7bit"` CTE**  
   Verify content is written verbatim (not base64-encoded).

10. **Interaction with `ForceQuotedPrintableCte` Encoder option**  
    Verify that `ContentTransferEncoding` on the part takes precedence over the encoder-level option.

## Interaction with existing features

| Feature | Interaction |
|---|---|
| `Encoder.ForceQuotedPrintableCte` | `ContentTransferEncoding` on the part takes precedence (more specific) |
| `TextContent()` gate | Bypassed when `ContentTransferEncoding` is set — this is the whole point |
| `ContentReader` (streaming) | If forced CTE ≠ base64, reader is slurped into `Content`; if base64, streaming path is unchanged |
| `rawContent` parser option | `rawContent` bypasses `setupMIMEHeaders()` entirely, so no interaction |
| `message/*` RFC 1341 special case | `ContentTransferEncoding` override takes precedence; caller is responsible for RFC compliance |
| Parsed parts (round-trip) | `ContentTransferEncoding` is a new field, defaults to `""`, so parsed parts behave exactly as before |

## Files to modify

| File | Change |
|---|---|
| `part.go` | Add `ContentTransferEncoding string` field; add `SetContentTransferEncoding()` method |
| `encode.go` | Add `resolveForcedCTE()`; modify `setupMIMEHeaders()` and `Encode()` to respect the field; handle `ContentReader` + non-base64 CTE |
| `encode_test.go` | Add test cases listed above |

## Out of scope

- Changing the automatic detection heuristics for non-text types (could be a separate enhancement)
- Adding a `MailBuilder`-level convenience method for setting CTE on attachments
- Validation that content actually fits the forced encoding (e.g., non-ASCII in 7bit) — follows Python's precedent of raising errors here, but we'll defer that
