# Documentation dependency maintenance

Docusaurus packages are updated together. The lockfile and explicit security overrides are required for reproducible builds; use `npm ci`, `npm run test:security`, and `npm run build` when validating updates.

The `image-size` override is an exact third-party replacement, `@nous-research/image-size@2.0.3`. The original package has no published fix for GHSA-w3rx-r6r6-pgpr and GHSA-5p2g-fcmc-qvqq. The replacement preserves the MIT license, supported formats and the existing CJS/ESM/fromFile APIs, and adds bounded ICNS/HEIF/JXL parser loops. Its published Git head is `e6dbb45e44b82eedbb81d150a31e4850dcb500d7`, following [the reviewed parser patch](https://github.com/NousResearch/image-size/commit/5d4dbf78a00e0390b0a143d6733d328395f1fae0). The lockfile pins its registry integrity. This is a maintained-fork trust decision; the original package itself is not being described as patched.

The remaining exact overrides update constrained transitive dependencies while retaining their callers' APIs: serialize-javascript 7.1.1 for webpack serialization, uuid 11.1.1 for SockJS v4 identifiers, and qs 6.16.0 for Express/body-parser query parsing. The website requires Node 20 or newer; CI verifies current Node 22 and 24 tooling.

Tests cover real PNG/JPEG/SVG assets through buffer/file APIs, bounded malformed-image regressions, and the callers' CJS contracts. Audit results and dependency aliases must be inspected together: no advisory ignore or audit exclusion is used. Reassess these overrides when Docusaurus adopts suitable patched dependencies upstream.
