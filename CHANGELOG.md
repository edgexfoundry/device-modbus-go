
<a name="EdgeX Modbus Device Service (found in device-modbus-go) Changelog"></a>
## EdgeX Modbus Device Service
[Github repository](https://github.com/edgexfoundry/device-modbus-go)

### Change Logs for EdgeX Dependencies
- [device-sdk-go](https://github.com/edgexfoundry/device-sdk-go/blob/main/CHANGELOG.md)
- [go-mod-core-contracts](https://github.com/edgexfoundry/go-mod-core-contracts/blob/main/CHANGELOG.md)
- [go-mod-bootstrap](https://github.com/edgexfoundry/go-mod-bootstrap/blob/main/CHANGELOG.md) (indirect dependency)
- [go-mod-messaging](https://github.com/edgexfoundry/go-mod-messaging/blob/main/CHANGELOG.md) (indirect dependency)
- [go-mod-registry](https://github.com/edgexfoundry/go-mod-registry/blob/main/CHANGELOG.md)  (indirect dependency)
- [go-mod-secrets](https://github.com/edgexfoundry/go-mod-secrets/blob/main/CHANGELOG.md) (indirect dependency)
- [go-mod-configuration](https://github.com/edgexfoundry/go-mod-configuration/blob/main/CHANGELOG.md) (indirect dependency)

## [v3.0.0] Minnesota - 2023-05-31 (Only compatible with the 3.x releases)

### Features ✨
- Add device validation function ([#452](https://github.com/edgexfoundry/device-modbus-go/pull/452))
    ```text
    BREAKING CHANGE: Implement `ValidateDevice` function to validate device protocol properties for core-metadata  
    ```
- Update for common config ([#413](https://github.com/edgexfoundry/device-modbus-go/pull/413))
    ```text
    BREAKING CHANGE: Configuration file is changed to remove common config settings
    ```
- Use latest SDK for MessageBus Request API ([#408](https://github.com/edgexfoundry/device-modbus-go/pull/408))
    ```text
    BREAKING CHANGE: Commands via MessageBus topic configuration are changed
    ```
- Remove ZeroMQ MessageBus capability ([#397](https://github.com/edgexfoundry/device-modbus-go/pull/397))
    ```text
    BREAKING CHANGE: ZeroMQ MessageBus capability no longer available
    ```

### Bug Fixes 🐛
- Fix protocol properties parsing error ([#261bb0a](https://github.com/edgexfoundry/device-modbus-go/commits/261bb0a))
- **snap:** Refactor to avoid conflicts with readonly config provider directory ([#437](https://github.com/edgexfoundry/device-modbus-go/issues/437)) ([#9bc48e7](https://github.com/edgexfoundry/device-modbus-go/commits/9bc48e7))

### Code Refactoring ♻
- Use integer for minimum and maximum properties ([#450](https://github.com/edgexfoundry/device-modbus-go/pull/450))
    ```text
    BREAKING CHANGE: Use integer for minimum and maximum properties
    ```
- Change configuration and devices files format to YAML ([#445](https://github.com/edgexfoundry/device-modbus-go/pull/445))
    ```text
    BREAKING CHANGE: Configuration files are now in YAML format, Default file name is now configuration.yaml
    ```
- Change protocol properties to use typed values ([#422](https://github.com/edgexfoundry/device-modbus-go/pull/422))
    ```text
    BREAKING CHANGE:
    - Update `Port`, `UnitID`, `Timeout`, `IdleTimeout` of modbus-tcp properties from string to integer
    - Update `UnitID`, `BaudRate`, `DataBits`, `StopBits`, `Timeout`, `IdleTimeout` of modbus-rtu properties from string to integer
    ```
- **snap:** Update command and metadata sourcing ([#432](https://github.com/edgexfoundry/device-modbus-go/issues/432)) ([#d976059](https://github.com/edgexfoundry/device-modbus-go/commits/d976059))
- **snap:** Drop the support for legacy snap env options ([#400](https://github.com/edgexfoundry/device-modbus-go/issues/400))
    ```text
    BREAKING CHANGE:
    - Drop the support for legacy snap options with env. prefix
    - Upgrade edgex-snap-hooks to v3
    - Upgrade edgex-snap-testing Github action to v3
    - Add snap's Go module to dependabot
    - Other minor refactoring
    ```

### Documentation 📖
- Add main branch Warning ([#478](https://github.com/edgexfoundry/device-modbus-go/issues/478)) ([#5189b6b](https://github.com/edgexfoundry/device-modbus-go/commits/5189b6b))

### Build 👷
- Update to Go 1.20, Alpine 3.17 and linter v1.51.2 ([#426](https://github.com/edgexfoundry/device-modbus-go/issues/426)) ([#7c5acbb](https://github.com/edgexfoundry/device-modbus-go/commits/7c5acbb))

## [v2.3.0] Levski - 2022-11-09  (Only compatible with the 2.x releases)

### Features ✨

- Add Service Metrics configuration  ([#387](https://github.com/edgexfoundry/device-modbus-go/issues/387)) ([#aaadd7f](https://github.com/edgexfoundry/device-modbus-go/commits/aaadd7f))
- Add NATS configuration and build option ([#376](https://github.com/edgexfoundry/device-modbus-go/issues/376)) ([#6ac2f22](https://github.com/edgexfoundry/device-modbus-go/commits/6ac2f22))
- Add commanding via message configuration ([#18fc891](https://github.com/edgexfoundry/device-modbus-go/commits/18fc891))
- Add go-winio to attribution (new SPIFFE dependency) ([#348](https://github.com/edgexfoundry/device-modbus-go/issues/348)) ([#20ae4d4](https://github.com/edgexfoundry/device-modbus-go/commits/20ae4d4))
- **snap:** add config interface with unique identifier ([#382](https://github.com/edgexfoundry/device-modbus-go/issues/382)) ([#9ccf6e7](https://github.com/edgexfoundry/device-modbus-go/commits/9ccf6e7))

### Code Refactoring ♻

- prefer spf13/cast to handle attributes ([#07d3dcc](https://github.com/edgexfoundry/device-modbus-go/commits/07d3dcc))
- **snap:** edgex-snap-hooks related upgrade ([#363](https://github.com/edgexfoundry/device-modbus-go/issues/363)) ([#614ae84](https://github.com/edgexfoundry/device-modbus-go/commits/614ae84))

### Build 👷

- Upgrade to Go 1.18 and optimize attribution script ([#361](https://github.com/edgexfoundry/device-modbus-go/issues/361)) ([#9739056](https://github.com/edgexfoundry/device-modbus-go/commits/9739056))

## [v2.2.0] Kamakura - 2022-05-11  (Only compatible with the 2.x releases)

### Features ✨
- Enable security hardening ([#106969e](https://github.com/edgexfoundry/device-modbus-go/commits/106969e))

### Bug Fixes 🐛
- **snap:** expose parent directory in device-config plug ([#013c32e](https://github.com/edgexfoundry/device-modbus-go/commits/013c32e))

### Code Refactoring ♻
- **snap:** remove redundant content indentifier ([#8c7365f](https://github.com/edgexfoundry/device-modbus-go/commits/8c7365f))

### Documentation 📖
- Update README.md for Modbus RTU usage ([#24e7f3a](https://github.com/edgexfoundry/device-modbus-go/commits/24e7f3a))

### Build 👷
- Update to latest SDK w/o ZMQ on windows ([#301d6ea](https://github.com/edgexfoundry/device-modbus-go/commits/301d6ea))
    ```
    BREAKING CHANGE:
    ZeroMQ no longer supported on native Windows for EdgeX
    MessageBus
    ```
- **deps:** Update device-sdk-go ([#16423cc](https://github.com/edgexfoundry/device-modbus-go/commits/16423cc))
- **snap:** source metadata from central repo ([#e14069c](https://github.com/edgexfoundry/device-modbus-go/commits/e14069c))

### Continuous Integration 🔄
- gomod changes related for Go 1.17 ([#864adf6](https://github.com/edgexfoundry/device-modbus-go/commits/864adf6))
- Go 1.17 related changes ([#27e7a32](https://github.com/edgexfoundry/device-modbus-go/commits/27e7a32))

## [v2.1.0] Jakarta - 2021-11-18  (Only compatible with the 2.x releases)

### Features ✨
- Update configuration for new CORS and Secrets File settings ([#d233c32](https://github.com/edgexfoundry/device-modbus-go/commits/d233c32))
- Supply string type ([#262](https://github.com/edgexfoundry/device-modbus-go/issues/262)) ([#c8e345b](https://github.com/edgexfoundry/device-modbus-go/commits/c8e345b))

### Bug Fixes 🐛
- Update all TOML to use quote and not single-quote ([#8c4c380](https://github.com/edgexfoundry/device-modbus-go/commits/8c4c380))

### Documentation 📖
- Update build status badge ([#f4dca52](https://github.com/edgexfoundry/device-modbus-go/commits/f4dca52))
- Update README.md and device profile ([#b6e2cd2](https://github.com/edgexfoundry/device-modbus-go/commits/b6e2cd2))
- **snap:** Update snap section in README ([#0ef4a91](https://github.com/edgexfoundry/device-modbus-go/commits/0ef4a91))
- **snap:** Update snap/README ([#54efd36](https://github.com/edgexfoundry/device-modbus-go/commits/54efd36))
- **snap:** Update snap section in README.md ([#fd029c8](https://github.com/edgexfoundry/device-modbus-go/commits/fd029c8))
- **snap:** Add snap section in README ([#8840200](https://github.com/edgexfoundry/device-modbus-go/commits/8840200))

### Build 👷
- Update to use released SDK ([#ce152de](https://github.com/edgexfoundry/device-modbus-go/commits/ce152de))
- Update to latest SDK and released go-mods ([#fd5a279](https://github.com/edgexfoundry/device-modbus-go/commits/fd5a279))
- Update to latest SDK ([#f129154](https://github.com/edgexfoundry/device-modbus-go/commits/f129154))
- update alpine base to 3.14 ([#037866b](https://github.com/edgexfoundry/device-modbus-go/commits/037866b))
- **snap:** upgrade base to core20 ([#494e11b](https://github.com/edgexfoundry/device-modbus-go/commits/494e11b))

### Continuous Integration 🔄
- Remove need for CI specific Dockerfile ([#d2a2473](https://github.com/edgexfoundry/device-modbus-go/commits/d2a2473))

## [v2.0.0] Ireland - 2021-06-30  (Only compatible with the 2.x releases)

### Features ✨
- Enable using MessageBus as the default ([#9743b62](https://github.com/edgexfoundry/device-modbus-go/commits/9743b62))
- Add secure MessagBus capability ([#aad8d66](https://github.com/edgexfoundry/device-modbus-go/commits/aad8d66))
- Use zero-based startingAddress ([#1219b9d](https://github.com/edgexfoundry/device-modbus-go/commits/1219b9d))
- Add Timeout and IdleTimeout to Protocol property ([#fc248eb](https://github.com/edgexfoundry/device-modbus-go/commits/fc248eb))
- Remove Logging configuration ([#b5c1d0b](https://github.com/edgexfoundry/device-modbus-go/commits/b5c1d0b))
### Bug Fixes 🐛
- Added missing InsecureSecrets section ([#0476c29](https://github.com/edgexfoundry/device-modbus-go/commits/0476c29))
### Code Refactoring ♻
- remove unimplemented InitCmd/RemoveCmd configuraiton ([#ebe2707](https://github.com/edgexfoundry/device-modbus-go/commits/ebe2707))
- Change PublishTopicPrefix value to be 'edgex/events/device' ([#8f6bb82](https://github.com/edgexfoundry/device-modbus-go/commits/8f6bb82))
- Update configuration for change to common ServiceInfo struct ([#64389da](https://github.com/edgexfoundry/device-modbus-go/commits/64389da))
    ```
    BREAKING CHANGE:
    Service configuration has changed
    ```
- Update to assign and uses new Port Assignments ([#462e2b0](https://github.com/edgexfoundry/device-modbus-go/commits/462e2b0))
    ```
    BREAKING CHANGE:
    Device Modbus default port number has changed to 59901
    ```
- rename example device AutoEvent Frequency to Interval ([#e2b33fc](https://github.com/edgexfoundry/device-modbus-go/commits/e2b33fc))
- Added go mod tidy to Docker and Makefile ([#0417c27](https://github.com/edgexfoundry/device-modbus-go/commits/0417c27))
- Update for new service key names and overrides for hyphen to underscore ([#be50c87](https://github.com/edgexfoundry/device-modbus-go/commits/be50c87))
    ```
    BREAKING CHANGE:
    Service key names used in configuration have changed.
    ```
- use v2 device-sdk ([#222](https://github.com/edgexfoundry/device-modbus-go/issues/222)) ([#7442c44](https://github.com/edgexfoundry/device-modbus-go/commits/7442c44))
### Documentation 📖
- Add badges to readme ([#6adafa8](https://github.com/edgexfoundry/device-modbus-go/commits/6adafa8))
### Build 👷
- update Dockerfiles to use go 1.16 ([#84e0467](https://github.com/edgexfoundry/device-modbus-go/commits/84e0467))
- **snap:** update snap v2 support ([#93f9e78](https://github.com/edgexfoundry/device-modbus-go/commits/93f9e78))
- **snap:** update go to 1.16 ([#8bae537](https://github.com/edgexfoundry/device-modbus-go/commits/8bae537))
### Continuous Integration 🔄
- update local docker image names ([#3113a28](https://github.com/edgexfoundry/device-modbus-go/commits/3113a28))
- fix link to Contributing.md ([#221](https://github.com/edgexfoundry/device-modbus-go/issues/221)) ([#bdd9c76](https://github.com/edgexfoundry/device-modbus-go/commits/bdd9c76))

<a name="v1.3.1"></a>
## [v1.3.1] - 2021-02-02
### Features ✨
- Enhance Modbus simulator to count reading amount ([#03c230d](https://github.com/edgexfoundry/device-modbus-go/commits/03c230d))
- **snap:** Add startup configure options ([#2f23f6b](https://github.com/edgexfoundry/device-modbus-go/commits/2f23f6b))
### Build 👷
- Upgrade sdk version to 1.4.0 ([#fd8f892](https://github.com/edgexfoundry/device-modbus-go/commits/fd8f892))
- update device-sdk-go to v1.3.1-dev.4 ([#24a4c3f](https://github.com/edgexfoundry/device-modbus-go/commits/24a4c3f))
- **deps:** Bump github.com/edgexfoundry/device-sdk-go ([#0c5bc27](https://github.com/edgexfoundry/device-modbus-go/commits/0c5bc27))
### Continuous Integration 🔄
- add semantic.yml for commit linting, update PR template to latest ([#ad0b01e](https://github.com/edgexfoundry/device-modbus-go/commits/ad0b01e))
- standardize dockerfiles ([#5a943b0](https://github.com/edgexfoundry/device-modbus-go/commits/5a943b0))

<a name="v1.3.0"></a>
## [v1.3.0] - 2020-11-18
### Features ✨
- Modify Modbus simulator to support device scaling ([#ef894a2](https://github.com/edgexfoundry/device-modbus-go/commits/ef894a2))
### Bug Fixes 🐛
- local snap development ([#9228997](https://github.com/edgexfoundry/device-modbus-go/commits/9228997))
- Lock Modbus TCP address with IP and Port ([#dec6cac](https://github.com/edgexfoundry/device-modbus-go/commits/dec6cac))
### Code Refactoring ♻
- Upgrade SDK to v1.2.4-dev.34 ([#7d26d0a](https://github.com/edgexfoundry/device-modbus-go/commits/7d26d0a))
- update dockerfile to appropriately use ENTRYPOINT and CMD, closes[#163](https://github.com/edgexfoundry/device-modbus-go/issues/163) ([#fb1cdd4](https://github.com/edgexfoundry/device-modbus-go/commits/fb1cdd4))
### Build 👷
- Upgrade to Go1.15 ([#e2ee2c1](https://github.com/edgexfoundry/device-modbus-go/commits/e2ee2c1))
- **all:** Enable use of DependaBot to maintain Go dependencies ([#4e643a1](https://github.com/edgexfoundry/device-modbus-go/commits/4e643a1))

<a name="v1.2.2"></a>
## [v1.2.2] - 2020-08-19
### Snap
- add env override for ProfilesDir ([#d854ea6](https://github.com/edgexfoundry/device-modbus-go/commits/d854ea6))
### Bug Fixes 🐛
- Fix swap operation for float32 dataType ([#1739004](https://github.com/edgexfoundry/device-modbus-go/commits/1739004))
### Code Refactoring ♻
- upgarde SDK to v1.2.0 ([#58041e0](https://github.com/edgexfoundry/device-modbus-go/commits/58041e0))
### Documentation 📖
- Add standard PR template ([#2944f8a](https://github.com/edgexfoundry/device-modbus-go/commits/2944f8a))
