# third_party

Local patched copies of upstream Go modules, wired in via `replace` directives in
`../go.mod`. Keep these in sync when merging from upstream.

## magic (github.com/liamg/magic v0.0.1)

Pulled in indirectly by `github.com/fiatjaf/khatru`'s blossom server, which uses
`magic.Lookup()` on the first 50 bytes of an upload to decide the blob's file
extension (`khatru/blossom/handlers.go`, `handleUpload`).

**Bug:** upstream's signature table contained an entry for "PalmPilot
Database/Document File" (ext `pdb`) whose magic was **24 zero bytes at offset 11**.
That is not a usable file signature -- it matches any file with a 24-byte zero run
at that position.

Every MP3 *without* an ID3v2 tag matches it: such a file starts with a bare
`0xFF 0xFB` frame header followed by a zero-padded Xing/Info header, so bytes
11..34 are all zeros. Because the `pdb` entry sits earlier in the table than the
mp3 entries, and `Lookup()` dispatches the table to concurrent workers in order
and returns the first match reported, `pdb` won every time (measured 2000/2000).

khatru then did `mime.TypeByExtension(".pdb")`, which on Ubuntu resolves via
`/etc/mime.types` to `chemical/x-pdb`, stored that in the blob's kind-24242 index
event `type` tag, and served it as `Content-Type` on every subsequent GET.
Result: MP3s uploaded through Blossom would not play in browsers.

**Patch:** the `pdb` entry is removed from `types.go` (see the comment left in its
place). It was the only all-zero signature in the table.

### Notes for future maintenance

- Do **not** "fix" this by switching khatru to `magic.LookupSync()`. That code path
  passes offset `0` for every signature instead of the signature's declared
  `Offset`, so all offset-based signatures silently stop matching. It happens to
  return `mp3` here, but it would break other formats.
- Do **not** try to bump to `liamg/magic` v1.0.1. It is a full API rewrite
  (`FileType.RecommendedExtension` / `.MIME` instead of `.Extension`) and requires
  Go >= 1.25, so it will not compile against khatru v0.19.1.
- `magic` returning *no* match is safe: khatru falls back to the client's declared
  `Content-Type` header. The danger is only a confidently wrong match. (This is
  why MP4 uploads were unaffected -- this table has no matching mp4 signature, so
  they always took the header fallback.)
- Blob types are cached. `EventStoreBlobIndexWrapper.Keep()` only writes a
  descriptor if one does not already exist for that sha256, so re-uploading a blob
  does **not** refresh a bad `type`. A blob stored with the wrong type must be
  DELETEd (which removes the kind-24242 event) and then re-uploaded. Since the
  bytes are unchanged the sha256 and URL stay the same.

Applied 2026-09-20.
