
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

## [v2.2.0] Kamakura - 2022-05-11  (Not Compatible with 1.x releases)

### Features ✨
- **all:** Make support services include spiffe in buildtime from Makefile ([#3969](https://github.com/edgexfoundry/edgex-go/issues/3969)) ([#efde56b5](https://github.com/edgexfoundry/edgex-go/commits/efde56b5))
- **all:** Location of client service obtained from the registry ([#3879](https://github.com/edgexfoundry/edgex-go/issues/3879)) ([#2ec8c03a](https://github.com/edgexfoundry/edgex-go/commits/2ec8c03a))
- **all:** Implement service metrics for Events and Readings persisted ([#3922](https://github.com/edgexfoundry/edgex-go/issues/3922)) ([#e06225a2](https://github.com/edgexfoundry/edgex-go/commits/e06225a2))
- **all:** Create Common DTOs with ServiceName ([#3870](https://github.com/edgexfoundry/edgex-go/issues/3870)) ([#d6b89ac6](https://github.com/edgexfoundry/edgex-go/commits/d6b89ac6))
- **data:** Make MaxEventSize a service configuration setting ([#3891](https://github.com/edgexfoundry/edgex-go/issues/3891)) ([#de3e46c9](https://github.com/edgexfoundry/edgex-go/commits/de3e46c9))
- **metadata:** Implement Writable.ProfileChange configurations ([#a20eafb3](https://github.com/edgexfoundry/edgex-go/commits/a20eafb3))
- **metadata:** Implement PATCH deviceprofile/resource API ([#a40e21f6](https://github.com/edgexfoundry/edgex-go/commits/a40e21f6))
- **metadata:** Implement DELETE device command API ([#ac0e4352](https://github.com/edgexfoundry/edgex-go/commits/ac0e4352))
- **metadata:** Implement DELETE device resource API ([#691b12bf](https://github.com/edgexfoundry/edgex-go/commits/691b12bf))
- **metadata:** Implement PATCH /deviceprofile/deviceCommand API ([#0fa976f9](https://github.com/edgexfoundry/edgex-go/commits/0fa976f9))
- **metadata:** Invoke validation before adding/updating device ([#3900](https://github.com/edgexfoundry/edgex-go/issues/3900)) ([#b3afc0ae](https://github.com/edgexfoundry/edgex-go/commits/b3afc0ae))
- **metadata:** Implement Post device command API ([#dcb0ece3](https://github.com/edgexfoundry/edgex-go/commits/dcb0ece3))
- **metadata:** Implement PATCH device profile basic info API ([#243ad5ab](https://github.com/edgexfoundry/edgex-go/commits/243ad5ab))
- **metadata:** Implement POST device resource ([#3930](https://github.com/edgexfoundry/edgex-go/issues/3930)) ([#7587fe9d](https://github.com/edgexfoundry/edgex-go/commits/7587fe9d))
- **security:** Add secret store config for delayed start ([#3948](https://github.com/edgexfoundry/edgex-go/issues/3948)) ([#6b188fc4](https://github.com/edgexfoundry/edgex-go/commits/6b188fc4))
- **security:** Enable security-hardened go binaries for cgo flags ([#3893](https://github.com/edgexfoundry/edgex-go/issues/3893)) ([#7a573170](https://github.com/edgexfoundry/edgex-go/commits/7a573170))
- **security:** Implement spiffe token provider services ([#3897](https://github.com/edgexfoundry/edgex-go/issues/3897)) ([#81bad761](https://github.com/edgexfoundry/edgex-go/commits/81bad761))
- **security:** Update the pre-authorized service list for spire server config ([#3942](https://github.com/edgexfoundry/edgex-go/issues/3942)) ([#c716e684](https://github.com/edgexfoundry/edgex-go/commits/c716e684))
- **security:** Use conditional go build tags for delayed start and no_messagebus ([#3967](https://github.com/edgexfoundry/edgex-go/issues/3967)) ([#0af16247](https://github.com/edgexfoundry/edgex-go/commits/0af16247))
- **snap:** Add secretstore tokens for edgex-ekuiper ([#3888](https://github.com/edgexfoundry/edgex-go/issues/3888)) ([#d0bb8bce](https://github.com/edgexfoundry/edgex-go/commits/d0bb8bce))
- **snap:** Add additional tokens for app-service-configurable profiles ([#3825](https://github.com/edgexfoundry/edgex-go/issues/3825)) ([#23881e65](https://github.com/edgexfoundry/edgex-go/commits/23881e65))

### Bug Fixes 🐛
- **all:** Correct jwt error when reading private key ([#3843](https://github.com/edgexfoundry/edgex-go/issues/3843)) ([#1876cd19](https://github.com/edgexfoundry/edgex-go/commits/1876cd19))
- **all:** Return 416 code when count is zero and offset > count ([#2c16b7b0](https://github.com/edgexfoundry/edgex-go/commits/2c16b7b0))
- **core-command:** Restore creation of DeviceServiceCommandClient in handler ([#89cc48a7](https://github.com/edgexfoundry/edgex-go/commits/89cc48a7))
- **metadata:** Fix the typo of notification message ([#5146f317](https://github.com/edgexfoundry/edgex-go/commits/5146f317))
- **metadata:** Not trigger DS callback if only lastConnected is updated ([#3976](https://github.com/edgexfoundry/edgex-go/issues/3976)) ([#7c172932](https://github.com/edgexfoundry/edgex-go/commits/7c172932))
- **metadata:** Add 404 case for not-yet-upgraded device-service ([#79655804](https://github.com/edgexfoundry/edgex-go/commits/79655804))
- **metadata:** Ignore validation result if device service unavailable ([#b9439831](https://github.com/edgexfoundry/edgex-go/commits/b9439831))
- **security:** Security-proxy-setup will not terminate on success ([#3917](https://github.com/edgexfoundry/edgex-go/issues/3917)) ([#d0a5aad2](https://github.com/edgexfoundry/edgex-go/commits/d0a5aad2))
- **snap:** Apply proxy's runtime config options after startup ([#3856](https://github.com/edgexfoundry/edgex-go/issues/3856)) ([#3825f82a](https://github.com/edgexfoundry/edgex-go/commits/3825f82a))
- **snap:** Deploy secrets token into separate mount points ([#3826](https://github.com/edgexfoundry/edgex-go/issues/3826)) ([#b0e5e643](https://github.com/edgexfoundry/edgex-go/commits/b0e5e643))

### Code Refactoring ♻
- **all:** Use formatted alternative log function instead of fmt.Sprintf ([#46183a66](https://github.com/edgexfoundry/edgex-go/commits/46183a66))

## [v2.1.0] Jakarta - 2021-11-18  (Not Compatible with 1.x releases)

### Features ✨
- **all:** Enable CORS headers for all services ([#3758](https://github.com/edgexfoundry/edgex-go/issues/3758)) ([#4fce4fd2](https://github.com/edgexfoundry/edgex-go/commits/4fce4fd2))
- **all:** Update multi instance response to have totalCount field ([#ea5a8f40](https://github.com/edgexfoundry/edgex-go/commits/ea5a8f40))
- **command:** Support object value type in Set Command ([#eaa9784a](https://github.com/edgexfoundry/edgex-go/commits/eaa9784a))
- **command:** Update Command V2 API to include totalCount field for MultiDeviceCoreCommandsResponse ([#4ad05991](https://github.com/edgexfoundry/edgex-go/commits/4ad05991))
- **data:** Implement V2 API to query readings by name and time range ([#3577](https://github.com/edgexfoundry/edgex-go/issues/3577)) ([#8a6c1010](https://github.com/edgexfoundry/edgex-go/commits/8a6c1010))
- **data:** New API to search Readings by multiple resource names ([#3766](https://github.com/edgexfoundry/edgex-go/issues/3766)) ([#e2d5d6cc](https://github.com/edgexfoundry/edgex-go/commits/e2d5d6cc))
- **data:** Enhance the performance of events deletion ([#3722](https://github.com/edgexfoundry/edgex-go/issues/3722)) ([#2de07aa5](https://github.com/edgexfoundry/edgex-go/commits/2de07aa5))
- **data:** Support Object value type in Reading ([#94769bcc](https://github.com/edgexfoundry/edgex-go/commits/94769bcc))
- **data:** Update MultiReadingsResponse to have totalCount field ([#07c09b9a](https://github.com/edgexfoundry/edgex-go/commits/07c09b9a))
- **data:** Update MultiEventsResponse to have totalCount field ([#d627eae0](https://github.com/edgexfoundry/edgex-go/commits/d627eae0))
- **data:** Implement new GET Readings API ([#1ef40f49](https://github.com/edgexfoundry/edgex-go/commits/1ef40f49))
- **metadata:** Send notification after updating device entity ([#3623](https://github.com/edgexfoundry/edgex-go/issues/3623)) ([#166d7917](https://github.com/edgexfoundry/edgex-go/commits/166d7917))
- **metadata:** Update Metadata V2 API to include totalCount field for multi-instance response ([#377c2adc](https://github.com/edgexfoundry/edgex-go/commits/377c2adc))
- **notifications:** Update Notification V2 API to include totalCount field ([#b1707c08](https://github.com/edgexfoundry/edgex-go/commits/b1707c08))
- **notifications:** Add new API to Get Transmissions by Notification id ([#3759](https://github.com/edgexfoundry/edgex-go/issues/3759)) ([#4de7b29e](https://github.com/edgexfoundry/edgex-go/commits/4de7b29e))
- **scheduler:** Validate Interval and IntervalAction before loading from config ([#3646](https://github.com/edgexfoundry/edgex-go/issues/3646)) ([#c934d262](https://github.com/edgexfoundry/edgex-go/commits/c934d262))
- **scheduler:** Update Scheduler V2 API to include totalCount field ([#2b972191](https://github.com/edgexfoundry/edgex-go/commits/2b972191))
- **security:** Add injection of Secure MessageBus creds for eKuiper connections ([#3778](https://github.com/edgexfoundry/edgex-go/issues/3778)) ([#fb769a00](https://github.com/edgexfoundry/edgex-go/commits/fb769a00))
- **security:** Add Secret File config setting ([#3788](https://github.com/edgexfoundry/edgex-go/issues/3788)) ([#adab5248](https://github.com/edgexfoundry/edgex-go/commits/adab5248))
- **security:** Enable modern cipher suite / TLSv1.3 only ([#3704](https://github.com/edgexfoundry/edgex-go/issues/3704)) ([#7380b5be](https://github.com/edgexfoundry/edgex-go/commits/7380b5be))
- **security:** Make Vault token TTL configurable ([#3675](https://github.com/edgexfoundry/edgex-go/issues/3675)) ([#19484f48](https://github.com/edgexfoundry/edgex-go/commits/19484f48))
- **snap:** Add vault ttl config support ([#ef3901f9](https://github.com/edgexfoundry/edgex-go/commits/ef3901f9))
- **snap:** Add additional devices to secret store lists in install hook ([#8ad81a0f](https://github.com/edgexfoundry/edgex-go/commits/8ad81a0f))

### Performance Improvements ⚡
- Change MaxResultCount setting to 1024 ([#8524b20a](https://github.com/edgexfoundry/edgex-go/commits/8524b20a))

### Bug Fixes 🐛
- **all:** Http response cannot be completed ([#3662](https://github.com/edgexfoundry/edgex-go/issues/3662)) ([#0ba6ba5b](https://github.com/edgexfoundry/edgex-go/commits/0ba6ba5b))
- **command:** Using the Device Service response code for Get Command ([#9f422825](https://github.com/edgexfoundry/edgex-go/commits/9f422825))
- **command:** Clean out database section from core command ([#0fae9ab3](https://github.com/edgexfoundry/edgex-go/commits/0fae9ab3))
- **command:** Fix core-command crashes error ([#86f6abfe](https://github.com/edgexfoundry/edgex-go/commits/86f6abfe))
- **data:** Add codes to remove db index of reading:deviceName:resourceName when deleting readings ([#173b0957](https://github.com/edgexfoundry/edgex-go/commits/173b0957))
- **metadata:** Remove operating state from device service ([#dc27294b](https://github.com/edgexfoundry/edgex-go/commits/dc27294b))
- **metadata:** Disable device notification by default ([#3789](https://github.com/edgexfoundry/edgex-go/issues/3789)) ([#c5f5ac19](https://github.com/edgexfoundry/edgex-go/commits/c5f5ac19))
- **metadata:** Device yaml marshal to Json  error ([#3683](https://github.com/edgexfoundry/edgex-go/issues/3683)) ([#e89d87e1](https://github.com/edgexfoundry/edgex-go/commits/e89d87e1))
- **metadata:** Add labels as part of query criteria when finding ([#3781](https://github.com/edgexfoundry/edgex-go/issues/3781)) ([#11dac8c4](https://github.com/edgexfoundry/edgex-go/commits/11dac8c4))
- **security:** Move JWT auth method to individual routes ([#3657](https://github.com/edgexfoundry/edgex-go/issues/3657)) ([#d2a5f5fe](https://github.com/edgexfoundry/edgex-go/commits/d2a5f5fe))
- **security:** Replace abandoned JWT package ([#3729](https://github.com/edgexfoundry/edgex-go/issues/3729)) ([#32c3a59f](https://github.com/edgexfoundry/edgex-go/commits/32c3a59f))
- **security:** Use localhost for kuiper config ([#8fa67b54](https://github.com/edgexfoundry/edgex-go/commits/8fa67b54))
- **security:** Secrets-config user connect using TLS ([#3698](https://github.com/edgexfoundry/edgex-go/issues/3698)) ([#258ae4e0](https://github.com/edgexfoundry/edgex-go/commits/258ae4e0))
- **security:** Remove unused curl executable from secretstore-setup Dockerfile - curl command executable is not used, so it is removed from the Docker file of service secretstore-setup ([#49239b82](https://github.com/edgexfoundry/edgex-go/commits/49239b82))
- **security:** Mismatched types int and int32 ([#3655](https://github.com/edgexfoundry/edgex-go/issues/3655)) ([#dbae55fc](https://github.com/edgexfoundry/edgex-go/commits/dbae55fc))
- **snap:** Rix app-rules-engine ([#651aaa83](https://github.com/edgexfoundry/edgex-go/commits/651aaa83))
- **snap:** Configure kuiper's REST service port ([#3770](https://github.com/edgexfoundry/edgex-go/issues/3770)) ([#a2b69b26](https://github.com/edgexfoundry/edgex-go/commits/a2b69b26))
- **snap:** Add kuiper message-bus config ([#602d7f53](https://github.com/edgexfoundry/edgex-go/commits/602d7f53))

### Code Refactoring ♻
- **all:** Clean up TOML quotes ([#3666](https://github.com/edgexfoundry/edgex-go/issues/3666)) ([#729eb473](https://github.com/edgexfoundry/edgex-go/commits/729eb473))
- **all:** Refactor io.Reader for reusing ([#3627](https://github.com/edgexfoundry/edgex-go/issues/3627)) ([#7434bcad](https://github.com/edgexfoundry/edgex-go/commits/7434bcad))
- **all:** Remove unused Redis client variables ([#905a639d](https://github.com/edgexfoundry/edgex-go/commits/905a639d))

## [v2.0.0] Ireland - 2021-06-30  (Not Compatible with 1.x releases)

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
