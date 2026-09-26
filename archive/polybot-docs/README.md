# Polybot Mintlify documents

`polybot-docs/` is the source of the real Mintlify site served at `https://polybot.yitwah.site/documents`.

## Update flow

1. Update or add research reports under `../docs/`.
2. Extend `docs.json` navigation when adding a report page.
3. Run `bash scripts/sync-research-reports.sh` and `node scripts/build-search-index.mjs` to copy the complete report files, their referenced figures/CSVs, and the offline search index.
4. Run `npm exec --yes --package=mintlify@4.2.693 -- mintlify validate`.
5. Run `mintlify export`, prefix the generated absolute links with `/documents/`, then atomically replace `/var/www/polybot-docs` on `vps-hk`.

The source reports are kept intact in the repository. The synchronization step makes only two MDX-compatibility changes to the published copy: it removes internal HTML marker comments and renders literal Python dictionaries as inline code. No research claim, table, or figure is shortened or rewritten during publication.
