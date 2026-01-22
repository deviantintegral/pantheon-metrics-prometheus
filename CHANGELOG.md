# Changelog

## [0.6.0](https://github.com/deviantintegral/pantheon-metrics-prometheus/compare/v0.5.0...v0.6.0) (2026-01-22)


### Features

* **grafana:** add account and plan filters to dashboard ([#79](https://github.com/deviantintegral/pantheon-metrics-prometheus/issues/79)) ([0083c8c](https://github.com/deviantintegral/pantheon-metrics-prometheus/commit/0083c8c1a116006fb82799a796ef56cf014aeca6))


### Bug Fixes

* **grafana:** average cache site query ([04688b1](https://github.com/deviantintegral/pantheon-metrics-prometheus/commit/04688b1d2d7e10248595234daf1685e5b3a2ea40))

## [0.5.0](https://github.com/deviantintegral/pantheon-metrics-prometheus/compare/v0.4.1...v0.5.0) (2026-01-22)


### ⚠ BREAKING CHANGES

* Metric names and labels have been renamed to follow Prometheus naming conventions from https://prometheus.io/docs/practices/naming/

### Features

* align metric and label naming with Prometheus conventions ([ce0e639](https://github.com/deviantintegral/pantheon-metrics-prometheus/commit/ce0e6393902a386e496ca991701790b057ff298a))
* **grafana:** update dashboard for new metric and label names ([bdb1e58](https://github.com/deviantintegral/pantheon-metrics-prometheus/commit/bdb1e5881de11c59f741c64e0e93d1e3f7fc9c8d))


### Bug Fixes

* **grafana:** use datasource variable for portability ([08196f2](https://github.com/deviantintegral/pantheon-metrics-prometheus/commit/08196f2e788f10d5e69ccfe6bd70101b8b2c7297))

## [0.4.1](https://github.com/deviantintegral/pantheon-metrics-prometheus/compare/v0.4.0...v0.4.1) (2026-01-22)


### Bug Fixes

* **ci:** use buildx imagetools to create multi-arch manifest ([3fc4b31](https://github.com/deviantintegral/pantheon-metrics-prometheus/commit/3fc4b315667c4b6039a25dde934da192141a81f6))

## [0.4.0](https://github.com/deviantintegral/pantheon-metrics-prometheus/compare/v0.3.0...v0.4.0) (2026-01-21)


### Features

* **docker:** add GHCR publishing for releases ([286e411](https://github.com/deviantintegral/pantheon-metrics-prometheus/commit/286e411ccde374ed9e30a147566e630e55a9a37b))

## [0.3.0](https://github.com/deviantintegral/pantheon-metrics-prometheus/compare/v0.2.1...v0.3.0) (2026-01-21)


### Features

* **collector:** use current request time for latest metric timestamp ([e528753](https://github.com/deviantintegral/pantheon-metrics-prometheus/commit/e5287530b9cea9f636da824ee56b88c0b50f79aa))

## [0.2.1](https://github.com/deviantintegral/pantheon-metrics-prometheus/compare/v0.2.0...v0.2.1) (2026-01-21)


### Bug Fixes

* **grafana:** single stat metric totals ([cac79f0](https://github.com/deviantintegral/pantheon-metrics-prometheus/commit/cac79f01da8af0e65eadef3783ca5505d81bfacc))

## [0.2.0](https://github.com/deviantintegral/pantheon-metrics-prometheus/compare/v0.1.0...v0.2.0) (2026-01-20)


### Features

* add -debug flag for HTTP request/response logging ([5b07be3](https://github.com/deviantintegral/pantheon-metrics-prometheus/commit/5b07be30d5ac40c0cdb9c64757b81f40ae78f17e))
* add -siteLimit flag to limit the number of sites queried ([7f27d3c](https://github.com/deviantintegral/pantheon-metrics-prometheus/commit/7f27d3c745d307f922b6bf197c87d364b6fc5caf))
* add Docker Compose setup with Prometheus and Grafana ([17f1173](https://github.com/deviantintegral/pantheon-metrics-prometheus/commit/17f11733ad48884475606815f621d6bda2ed0e0c))
* add dynamic version info to user agent header ([#50](https://github.com/deviantintegral/pantheon-metrics-prometheus/issues/50)) ([f1419b3](https://github.com/deviantintegral/pantheon-metrics-prometheus/commit/f1419b39f2d84d82521266d9dc290a3d3db040ba))
* add environment variable support for siteLimit and orgID ([2c339dc](https://github.com/deviantintegral/pantheon-metrics-prometheus/commit/2c339dc5b98086426590bd838729e90c37fa965f))
* add organization ID filter flag ([f78718f](https://github.com/deviantintegral/pantheon-metrics-prometheus/commit/f78718f55a9cd75d3de323b6b369c812f6ca52fd))
* **docker:** allow backfilling data ([03d432c](https://github.com/deviantintegral/pantheon-metrics-prometheus/commit/03d432c387e98ad818ba5fad8ba6fc9505ab4911))
* include organization sites in FetchAllSites ([a6f7072](https://github.com/deviantintegral/pantheon-metrics-prometheus/commit/a6f70727a48910f4df6638bda8dc0abc0170207c))
* update metrics incrementally during initial collection ([a1b2862](https://github.com/deviantintegral/pantheon-metrics-prometheus/commit/a1b286281f1e3dabce4360d44c3325b244668e07))


### Bug Fixes

* avoid duplicate site fetch on startup ([d0b83eb](https://github.com/deviantintegral/pantheon-metrics-prometheus/commit/d0b83eba16e4ee4c716e49b8a5eba338d0c8e8b6))
* **deps:** update module github.com/deviantintegral/terminus-golang to v0.7.0 ([#56](https://github.com/deviantintegral/pantheon-metrics-prometheus/issues/56)) ([1cd840d](https://github.com/deviantintegral/pantheon-metrics-prometheus/commit/1cd840dde65f7122b1c451700bb91e5e4d0ffdfc))
* handle "--" cache hit ratio gracefully ([#55](https://github.com/deviantintegral/pantheon-metrics-prometheus/issues/55)) ([cabddcb](https://github.com/deviantintegral/pantheon-metrics-prometheus/commit/cabddcb64d32da87414e9d3aadf6b467831581f6))
* initialize account token map before starting refresh ([96fbf24](https://github.com/deviantintegral/pantheon-metrics-prometheus/commit/96fbf2453d8f76ae2198662a1e547bcd97498b6f))
* make code coverage annotation step always run ([52b962a](https://github.com/deviantintegral/pantheon-metrics-prometheus/commit/52b962a543de9fce3adcf9bc85dca160182f1728))
* mark getAccountID as not deprecated ([#31](https://github.com/deviantintegral/pantheon-metrics-prometheus/issues/31)) ([34e1b61](https://github.com/deviantintegral/pantheon-metrics-prometheus/commit/34e1b6132019385d17c6ac8defcf4be9b666ddc4))
* resolve goreleaser deprecation warnings ([#42](https://github.com/deviantintegral/pantheon-metrics-prometheus/issues/42)) ([8b311bb](https://github.com/deviantintegral/pantheon-metrics-prometheus/commit/8b311bbef23d39484744977679298a3d43e78b9a))
* use 1-based site numbering in refresh log message ([afc0a0f](https://github.com/deviantintegral/pantheon-metrics-prometheus/commit/afc0a0ffb80a6731e342793a745956985ba8c260))
* use forked codecoverage action to handle large PRs ([fb23ea5](https://github.com/deviantintegral/pantheon-metrics-prometheus/commit/fb23ea50e8f5742d78b48713fb5c939efc604355))

## 0.1.0 (2026-01-16)


### Features

* add logging for site list updates and metrics refresh ([a02bc25](https://github.com/deviantintegral/pantheon-metrics-prometheus/commit/a02bc2536722c120c165e401e30ea5b69c6f6fcf))
* Populate site list before collecting metrics ([47ccd5b](https://github.com/deviantintegral/pantheon-metrics-prometheus/commit/47ccd5b41f87c4f53f881e9aa2298bb1340b756c))
* Start HTTP server immediately and collect metrics in background ([0848386](https://github.com/deviantintegral/pantheon-metrics-prometheus/commit/0848386022cd8b6e473e9fec237ebb8c6dab78b2))
* start HTTP server immediately and collect metrics in background ([#25](https://github.com/deviantintegral/pantheon-metrics-prometheus/issues/25)) ([09b0d37](https://github.com/deviantintegral/pantheon-metrics-prometheus/commit/09b0d37ec8493b99206debb619cb82306142e7fa))
* use email from terminus auth:whoami for account identification ([b920d42](https://github.com/deviantintegral/pantheon-metrics-prometheus/commit/b920d42d005440e2a1aa1585a02cc1fbc4fe4e9e))
* use email from terminus auth:whoami for account identification ([#27](https://github.com/deviantintegral/pantheon-metrics-prometheus/issues/27)) ([a4f0f25](https://github.com/deviantintegral/pantheon-metrics-prometheus/commit/a4f0f25f9810516bc8811b0fd1b295b1f2e80450))


### Bug Fixes

* disable config verification in golangci-lint-action ([583c670](https://github.com/deviantintegral/pantheon-metrics-prometheus/commit/583c670c21f956f8ee46656600f3acf35fbf456f))
* make TestRefreshMetricsWithQueueTickerFires actually validate ticker behavior ([0349232](https://github.com/deviantintegral/pantheon-metrics-prometheus/commit/03492325a8a124b2d81001ee70d36db484655f3d))
* make TestRefreshMetricsWithQueueTickerFires actually validate ticker behavior ([#26](https://github.com/deviantintegral/pantheon-metrics-prometheus/issues/26)) ([dea0eef](https://github.com/deviantintegral/pantheon-metrics-prometheus/commit/dea0eef0bbe6f55fb4ed9c16bbf8721683be4e28))
* resolve all remaining golangci-lint errors ([27162c7](https://github.com/deviantintegral/pantheon-metrics-prometheus/commit/27162c7e2fd50e38fd8abfb6f35f8da2118bd39e))
* resolve golangci-lint errors ([b602494](https://github.com/deviantintegral/pantheon-metrics-prometheus/commit/b60249490a89413a6aab0fba8845889b71d04b76))
* separate Deprecated notice into dedicated paragraph ([38b3a4c](https://github.com/deviantintegral/pantheon-metrics-prometheus/commit/38b3a4cb070fcae2b8619641139804cf33c2ce57))
* Update example-metrics.json format and parsing to include timeseries wrapper ([f1bb75e](https://github.com/deviantintegral/pantheon-metrics-prometheus/commit/f1bb75e7b762a0cddf9a498fff2f628e129709af))
* update example-metrics.json format and parsing to include timeseries wrapper ([#24](https://github.com/deviantintegral/pantheon-metrics-prometheus/issues/24)) ([1b171bd](https://github.com/deviantintegral/pantheon-metrics-prometheus/commit/1b171bd8d95637326829012ddfeb3ea8c6df4cf8))


### Miscellaneous Chores

* release as 0.1.0 ([6667ceb](https://github.com/deviantintegral/pantheon-metrics-prometheus/commit/6667ceb504d5f9e370560c3dac0dff38ea688430))
