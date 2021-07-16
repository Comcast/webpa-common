# Changelog
All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](http://keepachangelog.com/en/1.0.0/)
and this project adheres to [Semantic Versioning](http://semver.org/spec/v2.0.0.html).

## [Unreleased]
- Bumped argus version, removed dependency that couldn't be found.  Updated argus client-related metrics. [#582](https://github.com/xmidt-org/webpa-common/pull/582)

## [v1.11.8]
- Bumped bascule and argus versions. [#581](https://github.com/xmidt-org/webpa-common/pull/581)

## [v1.11.7]
- Deprecate basculechecks and basculemetrics packages. [#578](https://github.com/xmidt-org/webpa-common/pull/578)
- Fix bug around metric listener creation in basculemetrics. [#579](https://github.com/xmidt-org/webpa-common/pull/579)

## [v1.11.6]
- Use updated Argus client with OpenTelemetry integration. [#573](https://github.com/xmidt-org/webpa-common/pull/573) thanks to @utsavbatra5
- Add tracing to fanout client used by Scytale. [#576](https://github.com/xmidt-org/webpa-common/pull/576)

## [v1.11.5]
### Changed
- Update uber/fx measures setup-code for basculechecks and basculemetrics. [#558](https://github.com/xmidt-org/webpa-common/pull/558)
- Update code to use the latest Argus changes. [#557](https://github.com/xmidt-org/webpa-common/pull/557)
- Update item ID format for Argus. [#560](https://github.com/xmidt-org/webpa-common/pull/560)
- Add deprecation warning to xwebhook. [#564](https://github.com/xmidt-org/webpa-commmon/pull/564)
- Update Argus version to v0.3.12. [#567](https://github.com/xmidt-org/webpa-commmon/pull/567)

### Fixed
- Keep old metrics unchanged for basculemetrics for backward compatibility. [#560](https://github.com/xmidt-org/webpa-common/pull/560)

### Added
- Add ability to drain devices by specific parameters. [#554](https://github.com/xmidt-org/webpa-common/pull/554)
- Add filtering capability to gate device connections. [#547](https://github.com/xmidt-org/webpa-common/pull/547)

## [v1.11.4]
- Fixed string slice casting issue in basculechecks package. [#548](https://github.com/xmidt-org/webpa-common/pull/548)

## [v1.11.3]
### Changed
- Updated basculechecks library to run different capability checks based on the endpoint hit. [#538](https://github.com/xmidt-org/webpa-common/pull/538)

### Fixed
- Bug in which only mTLS was allowed as valid config for an HTTP server. [#544](https://github.com/xmidt-org/webpa-common/pull/544)

## [v1.11.2]
### Changed
- Update xwebhook to a stable version of argus. [#537](https://github.com/xmidt-org/webpa-common/pull/537)
- Obfuscate secrets when listing webhooks. [#534](https://github.com/xmidt-org/webpa-common/pull/534)
- Add client ID fallback method when adding webhook. [#533](https://github.com/xmidt-org/webpa-common/pull/533)

## [v1.11.1]
### Changed
- Made 6060 pprof server's default address. [#521](https://github.com/xmidt-org/webpa-common/pull/521)
- Bumped bascule version and made attributes-related changes. [#525](https://github.com/xmidt-org/webpa-common/pull/525)

## [v1.11.0]
### Changed
- Updated capability checker to be more modular and have more configurable checks. [#522](https://github.com/xmidt-org/webpa-common/pull/522)
- Exported function for determining partner ID to use in a metric label from a list of partner IDs. [#523](https://github.com/xmidt-org/webpa-common/pull/523)

## [v1.10.8]
### Changed
- Default partnerID in metadata changed from "" to "unknown". [#518](https://github.com/xmidt-org/webpa-common/pull/518)

### Fixed
- A typo where "monitor" was written as "monito" for a WRP source check. [#518](https://github.com/xmidt-org/webpa-common/pull/518)

## [v1.10.7]
### Added
- Added eventKey to service discovery metric. [#509](https://github.com/xmidt-org/webpa-common/pull/509)
- Added `trust` label to hardware_model gauge. [#512](https://github.com/xmidt-org/webpa-common/pull/512)

### Fixed
- Fix serviceEndpoints unit tests failing. [#510](https://github.com/xmidt-org/webpa-common/pull/510)
- Fix device Metadata Trust() method float64 casting to int. [#511](https://github.com/xmidt-org/webpa-common/pull/511)

### Changed
- Add configurable check for the source of inbound (device => cloud service) WRP messages. [#507](https://github.com/xmidt-org/webpa-common/pull/507)
- Populate empty inbound WRP.content_type field with `application/octet-stream`. [#508](https://github.com/xmidt-org/webpa-common/pull/508)
- Update inbound WRP source check logic. [#511](https://github.com/xmidt-org/webpa-common/pull/511)
- Standardize bascule measure factories. [#562](https://github.com/xmidt-org/webpa-common/pull/562)


## [v1.10.6]
### Fixed
- Fixed ServiceEndpoints not faning out when a datacenter is empty [#504](https://github.com/xmidt-org/webpa-common/pull/504)

## [v1.10.5]
### Fixed
- Fixed AllDataCenters config behavior and renamed it to CrossDatacenter. [#502](https://github.com/xmidt-org/webpa-common/pull/502)
- In service/consul, reset the last wait index if it goes backwards from the previous call. [#501](https://github.com/xmidt-org/webpa-common/pull/501)

## [v1.10.4]
### Fixed
- Added default port for creating a xresolver connection. [#499](https://github.com/xmidt-org/webpa-common/pull/499)

### Added
- Add partner and firmware label to existing `hardware_model` gauge. [#498](https://github.com/xmidt-org/webpa-common/pull/498)

### Changed
- Update word from blacklist to denylist. [#495](https://github.com/xmidt-org/webpa-common/pull/495)
- Update references to the main branch. [#497](https://github.com/xmidt-org/webpa-common/pull/497)


## [v1.10.3]
### Fixed
- Fixed xresolver failing to create route when default port is used. [#494](https://github.com/xmidt-org/webpa-common/pull/494)

## [v1.10.2]
### Fixed
- Fixed `ConsulWatch` in xresolver by storing and watching the correct part of the url. [#490](https://github.com/xmidt-org/webpa-common/pull/490)
- Fixed consul service discovery to pass QueryOptions. [#490](https://github.com/xmidt-org/webpa-common/pull/490)

### Changed 
- Device metadata implementation is now thread-safe and optimized for reads. [#489](https://github.com/xmidt-org/webpa-common/pull/489)


## [v1.10.1]
### Fixed 
- Device metadata didn't return a read-only view of its map claims resulting in data races. [#483](https://github.com/xmidt-org/webpa-common/pull/483)


## [v1.10.0]
### Removed
- Prune types package containing unnecessary custom time.Duration [#430](https://github.com/xmidt-org/webpa-common/pull/430)

### Fixed
- Fixed service discovery metrics label to only include actual service name [#480](https://github.com/xmidt-org/webpa-common/pull/480)
- Fixed rehasher issue in which it was not filtering out service discovery events from irrelevant services [#480](https://github.com/xmidt-org/webpa-common/pull/480)

## Deprecated
- Deprecate various packages [#443](https://github.com/xmidt-org/webpa-common/pull/443)


## [v1.9.0]
- Create simple component to work with a device's metadata [#447](https://github.com/xmidt-org/webpa-common/pull/447)

## [v1.8.1]
- change webhooks package to not use `logging` functions [#469](https://github.com/xmidt-org/webpa-common/pull/469)

## [v1.8.0]
- upgrade wrp-go version to v3.0.1 for subpackages convey and device [#460](https://github.com/xmidt-org/webpa-common/pull/460)
- Remove token logging from secure package [463](https://github.com/xmidt-org/webpa-common/pull/463)
- Increase reliability of travis unit tests by refactoring racy testLogger [#462](https://github.com/xmidt-org/webpa-common/pull/462)

## [v1.7.0]
- typo fix fallback content-type value [#457](https://github.com/xmidt-org/webpa-common/pull/457)
- changed how we determine the endpoint label for metrics [#459](https://github.com/xmidt-org/webpa-common/pull/459)

## [v1.6.3]
- added session-id to request response message [#454](https://github.com/xmidt-org/webpa-common/pull/454)
- added wrp-go v2.0.0 to device package [#454](https://github.com/xmidt-org/webpa-common/pull/454)

## [v1.6.2]
- Updated capabilityChecker to include metrics and configure whether or not to return errors [#449](https://github.com/xmidt-org/webpa-common/pull/449)

## [v1.6.1]
- Fixed panic from assignment to entry in nil map [#453](https://github.com/xmidt-org/webpa-common/pull/453)

## [v1.6.0]
- Added session-id to device information [#451](https://github.com/xmidt-org/webpa-common/pull/451)
- Added device metadata to outbound events [#451](https://github.com/xmidt-org/webpa-common/pull/451)

## [v1.5.1]
- Added automated releases using travis [#444](https://github.com/xmidt-org/webpa-common/pull/444)
- Bumped bascule version to v0.7.0 and updated basculechecks to match [#444](https://github.com/xmidt-org/webpa-common/pull/444)

## [v1.5.0]
- reduced logging for xhttp retry #441
- modified capabilities check to be more restrictive #440

## [v1.4.0]
- Moved from glide to go modules
- Updated bascule version to v0.5.0
- Updated wrp-go to v1.3.3
- Updated README to match go modules
- No longer accept retries in webhook.W

## [v1.3.2]
- Bump Bascule to v0.2.7

## [v1.3.1]
- Downgraded bascule version

## [v1.3.0]
- removed `wrp` package

## [v1.2.0]
- Added minVersion and maxVersion to `server` package.
- Added cpuprofile and memprofile flags.
- Updated import paths.


## [v1.1.0]
- Added ability to check status code for retrying http request
- Added ability to update http.Request for `xhttp` retry
- Added MaxRetry and AlternativeURLS for update webhook config

## [v1.0.1]
- Fix for https://github.com/Comcast/webpa-common/issues/364
- Removed unused dep files.
- Added capability checks to be used when consuming `bascule` package.
- Fix for responseRequest test that was intermittently failing.

## [v1.0.0]
 - The first official release. We will be better about documenting changes 
   moving forward.

[Unreleased]: https://github.com/xmidt-org/webpa-common/compare/v1.11.8...HEAD
[v1.11.8]: https://github.com/xmidt-org/webpa-common/compare/v1.11.7...v1.11.8
[v1.11.7]: https://github.com/xmidt-org/webpa-common/compare/v1.11.6...v1.11.7
[v1.11.6]: https://github.com/xmidt-org/webpa-common/compare/v1.11.5...v1.11.6
[v1.11.5]: https://github.com/xmidt-org/webpa-common/compare/v1.11.4...v1.11.5
[v1.11.4]: https://github.com/xmidt-org/webpa-common/compare/v1.11.3...v1.11.4
[v1.11.3]: https://github.com/xmidt-org/webpa-common/compare/v1.11.2...v1.11.3
[v1.11.2]: https://github.com/xmidt-org/webpa-common/compare/v1.11.1...v1.11.2
[v1.11.1]: https://github.com/xmidt-org/webpa-common/compare/v1.11.0...v1.11.1
[v1.11.0]: https://github.com/xmidt-org/webpa-common/compare/v1.10.8...v1.11.0
[v1.10.8]: https://github.com/xmidt-org/webpa-common/compare/v1.10.7...v1.10.8
[v1.10.7]: https://github.com/xmidt-org/webpa-common/compare/v1.10.6...v1.10.7
[v1.10.6]: https://github.com/xmidt-org/webpa-common/compare/v1.10.5...v1.10.6
[v1.10.5]: https://github.com/xmidt-org/webpa-common/compare/v1.10.4...v1.10.5
[v1.10.4]: https://github.com/xmidt-org/webpa-common/compare/v1.10.3...v1.10.4
[v1.10.3]: https://github.com/xmidt-org/webpa-common/compare/v1.10.2...v1.10.3
[v1.10.2]: https://github.com/xmidt-org/webpa-common/compare/v1.10.1...v1.10.2
[v1.10.1]: https://github.com/xmidt-org/webpa-common/compare/v1.10.0...v1.10.1
[v1.10.0]: https://github.com/xmidt-org/webpa-common/compare/v1.9.0...v1.10.0
[v1.9.0]: https://github.com/xmidt-org/webpa-common/compare/v1.8.1...v1.9.0
[v1.8.1]: https://github.com/xmidt-org/webpa-common/compare/v1.8.0...v1.8.1
[v1.8.0]: https://github.com/xmidt-org/webpa-common/compare/v1.7.0...v1.8.0
[v1.7.0]: https://github.com/xmidt-org/webpa-common/compare/v1.6.3...v1.7.0
[v1.6.3]: https://github.com/xmidt-org/webpa-common/compare/v1.6.2...v1.6.3
[v1.6.2]: https://github.com/xmidt-org/webpa-common/compare/v1.6.1...v1.6.2
[v1.6.1]: https://github.com/xmidt-org/webpa-common/compare/v1.6.0...v1.6.1
[v1.6.0]: https://github.com/xmidt-org/webpa-common/compare/v1.5.1...v1.6.0
[v1.5.1]: https://github.com/xmidt-org/webpa-common/compare/v1.5.0...v1.5.1
[v1.5.0]: https://github.com/xmidt-org/webpa-common/compare/v1.4.0...v1.5.0
[v1.4.0]: https://github.com/xmidt-org/webpa-common/compare/v1.3.2...v1.4.0
[v1.3.2]: https://github.com/xmidt-org/webpa-common/compare/v1.3.1...v1.3.2
[v1.3.1]: https://github.com/xmidt-org/webpa-common/compare/v1.3.0...v1.3.1
[v1.3.0]: https://github.com/xmidt-org/webpa-common/compare/v1.2.0...v1.3.0
[v1.2.0]: https://github.com/xmidt-org/webpa-common/compare/v1.1.0...v1.2.0
[v1.1.0]: https://github.com/xmidt-org/webpa-common/compare/v1.0.1...v1.1.0
[v1.0.1]: https://github.com/xmidt-org/webpa-common/compare/v1.0.0...v1.0.1
[v1.0.0]: https://github.com/xmidt-org/webpa-common/compare/v0.9.0-alpha...v1.0.0
