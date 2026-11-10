# v1.8.14 security maintenance

Based on the osinx fork tag v1.8.13; retains HTTPS proxy entry, Proxy Protocol, custom authentication and HTTP port handling.

- Backport upstream HTTP/2 RST_STREAM fix: use the default HTTP/2 write scheduler (upstream commits 1c172f72 and 58771e12).
- Backport upstream SOCKS5 dial failure fix: do not close a nil connection (388f47ad).
- Restore read-side-only shutdown for plain TCP when upstream closes, so a response is not truncated before it is flushed. Wrapped/TLS entries no longer close the client prematurely.
- Add regressions for SOCKS5 failure, HTTP/2 reset, and authenticated HTTP/HTTPS entries with and without Proxy Protocol. The HTTP/2 test is carried from upstream.
- Isolate certificate tests from the user's existing CA store and avoid external DNS in the unreachable-host test.
- Require Go 1.26.0 and select toolchain 1.26.10; release builds also use Go 1.26.10.
- Update x/net to 0.62.0, x/text to 0.43.0, Brotli to 1.2.7 and compress to 1.20.2, with associated x/sys dependency updates.

This is a focused maintenance release, not a merge of upstream v1.9.x features. A rebuilt executable is required to receive standard-library fixes; updating go.mod alone does not fix an existing deployment.
