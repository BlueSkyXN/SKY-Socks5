# Proxy Sources

The default source list lives at `configs/sources.txt`.

Each non-empty, non-comment line should be an HTTP or HTTPS URL that returns SOCKS5 proxy candidates. Newline-delimited lists are preferred, but whitespace-delimited records are also supported for sources that publish many candidates on one line. The parser accepts common line formats such as:

- `host:port`
- `socks5://host:port`
- `socks5://user:pass@host:port`
- `socks5://host:port,FR,Paris`
- `[ipv6-address]:port`

Blank lines and full-line `#` comments in fetched source responses are treated as structural text and skipped before candidate parsing.

## Current source policy

- Prefer SOCKS5-specific endpoints over mixed protocol lists.
- Prefer text endpoints over JSON. CSV-like records are accepted when the first field is the proxy URL and following fields provide source metadata such as country or city.
- Keep source additions small and reviewable.
- Before merging a default source addition, locally check only source fetchability, returned format, and parser compatibility. Do not use local SOCKS5 reachability as CD evidence; live proxy validation belongs in the CD workflow's network environment.
- Country-scoped Proxifly files can contain mixed protocols. The CLI accepts SOCKS5 candidates and rejects `http://` or `https://` proxy records instead of testing them as SOCKS5.
- Remove sources that consistently return HTTP errors such as `404 Not Found` or `502 Bad Gateway`; they should not stay in the default aggregate list.
- Avoid adding very large or mixed-protocol feeds to the default aggregate list until per-source CD results show useful unique and valid contribution.
- A high upstream candidate count is not the same as a high validated count. Validation requires an end-to-end SOCKS5 connection to the probe URL within the configured timeout and an accepted HTTP status.
- The main CSV intentionally keeps only fields useful for filtering and comparison. Full Cloudflare trace key/value data is preserved in `proxy_report.json`; see `docs/result-fields.md` for the field policy.

## May 2026 source review additions

The default list includes these smaller raw text sources from the May 2026 source review because they currently fetch successfully and return parser-compatible SOCKS5 candidate lines:

```text
https://raw.githubusercontent.com/ClearProxy/checked-proxy-list/main/socks5/raw/all.txt
https://raw.githubusercontent.com/jetkai/proxy-list/main/online-proxies/txt/proxies-socks5.txt
https://raw.githubusercontent.com/Thordata/awesome-free-proxy-list/main/proxies/socks5.txt
https://raw.githubusercontent.com/proxygenerator1/ProxyGenerator/main/Stable/socks5.txt
https://raw.githubusercontent.com/proxygenerator1/ProxyGenerator/main/MostStable/socks5.txt
https://raw.githubusercontent.com/vakhov/fresh-proxy-list/master/socks5.txt
```

Larger compatible feeds such as `r00tee/Proxy-List`, `dpangestuw/Free-Proxy`, `iplocate/free-proxy-list`, `vmheaven/VMHeaven.io-Free-Proxy-List`, and `SoliSpirit/proxy-list` are intentionally not enabled by default yet. They should be compared through CD per-source `unique_proxies` and `valid_proxies` results before becoming part of the default aggregate list. `monosans/proxy-list` is also kept out until its raw endpoint consistently returns a non-empty list, and `Sage520/Proxy-List` is kept out until CD shows non-zero valid contribution. Meta-source URL lists, such as `gfpcom/free-proxy-list` `sources/socks5.txt`, need nested-source support before they can be used directly.

## ProxyScrape

ProxyScrape is included through its v4 free-list API in paged text mode:

```text
https://api.proxyscrape.com/v4/free-proxy-list/get?request=getproxies&protocol=socks5&timeout=10000&country=all&skip=0&limit=2000
https://api.proxyscrape.com/v4/free-proxy-list/get?request=getproxies&protocol=socks5&timeout=10000&country=all&skip=2000&limit=2000
```

The previous v2 URL returned a valid text list, but a single request did not cover the full observed list. The default config therefore uses explicit `skip` and `limit` pages.

## Proxifly

Proxifly's SOCKS5 list is included through its raw GitHub text endpoint:

```text
https://raw.githubusercontent.com/proxifly/free-proxy-list/main/proxies/protocols/socks5/data.txt
```

This endpoint returns `socks5://host:port` lines, which are supported by the parser. The jsDelivr CDN URL is avoided in the default list because it can lag behind the raw repository content.

Proxifly country endpoints, for example `proxies/countries/FR/data.txt` or `data.csv`, are useful for country-specific smoke runs, but they are broader than the default SOCKS5 protocol feed. Use their SOCKS5 records when the goal is this repository's `validated_proxies.txt` and `validated_proxies.csv` outputs.
