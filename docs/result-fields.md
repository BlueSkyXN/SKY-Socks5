# Result Field Policy

SKY-Socks5 treats Cloudflare `/cdn-cgi/trace` as an exit-observation probe, not as a complete IP intelligence source. The CSV output should stay small and useful for filtering, while the JSON report keeps raw trace data for audit and future enrichment.

## Output design

`validated_proxies.csv` keeps fields that help compare validated proxies:

| Group | Fields | Purpose |
| --- | --- | --- |
| Entry endpoint | `proxy_address`, `entry_host`, `entry_port` | Shows the tested SOCKS5 entry point. |
| Source signal | `source_count`, `duplicate_count`, `source_urls`, `source_country`, `source_city` | Shows where the candidate came from and whether multiple upstreams repeated it. |
| Exit signal | `exit_ip`, `exit_country`, `entry_exit_same_ip`, `source_country_matches_exit` | Shows what Cloudflare observed after connecting through the proxy and whether it matches source metadata. |
| Cloudflare routing signal | `cloudflare_colo`, `cloudflare_http`, `cloudflare_flags` | Keeps compact routing and anomaly hints. `cloudflare_flags` only records non-default `warp`, `gateway`, or `rbi` values. |
| Validation evidence | `status_code`, `duration_ms`, `validated_at`, `probe_url`, `validation_timeout` | Makes the run reproducible and comparable. |

`proxy_report.json` keeps the same validated proxy records plus the full `cloudflare_trace` map. Diagnostic Cloudflare fields such as `tls`, `sni`, `kex`, `fl`, `sliver`, `uag`, and `ts` stay there instead of becoming top-level CSV columns.

## Process design

- CI runs Go program tests only.
- Build creates cross-platform binary archives only.
- CD runs live proxy validation in the GitHub Actions network environment and uploads generated result artifacts.
- Local development should use `go test`, `go vet`, `go build`, `actionlint`, and static artifact checks. Do not use local proxy-validation results as evidence for CD quality because network conditions differ.
- CD summaries should explain the input source mode, validation policy, output files, valid rate, exit-country distribution, Cloudflare colo distribution, source-country match rate, entry/exit IP agreement, timing percentiles, and any non-default Cloudflare flags.

Cloudflare trace does not provide ASN, ISP, city, region, or risk score. Add a separate IP enrichment layer if those fields become product requirements.
