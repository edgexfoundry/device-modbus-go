
<a name="EdgeX Modbus Device Service (found in device-modbus-go) Changelog"></a>
## EdgeX Modbus Device Service
[Github repository](https://github.com/edgexfoundry/device-modbus-go)

## [v2.0.0] Ireland - 2021-06-30  (Not Compatible with 1.x releases)
### Change Logs for EdgeX Dependencies
- [device-sdk-go](https://github.com/edgexfoundry/device-sdk-go/blob/v2.0.0/CHANGELOG.md)
- [go-mod-core-contracts](https://github.com/edgexfoundry/go-mod-core-contracts/blob/v2.0.0/CHANGELOG.md)

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
