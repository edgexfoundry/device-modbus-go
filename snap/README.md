# EdgeX Modbus Device Service Snap
[![snap store badge](https://raw.githubusercontent.com/snapcore/snap-store-badges/master/EN/%5BEN%5D-snap-store-black-uneditable.png)](https://snapcraft.io/edgex-device-modbus)

This folder contains snap packaging for the EdgeX Modbus Device Service Snap

The snap currently supports both `amd64` and `arm64` platforms.

## Installation

### Installing snapd
The snap can be installed on any system that supports snaps. You can see how to install
snaps on your system [here](https://snapcraft.io/docs/installing-snapd/6735).

However for full security confinement, the snap should be installed on an
Ubuntu 18.04 LTS or later (Desktop or Server), or a system running Ubuntu Core 18 or later.

### Installing EdgeX Device Modbus as a snap
The snap is published in the snap store at https://snapcraft.io/edgex-device-modbus.
You can see the current revisions available for your machine's architecture by running the command:

```bash
$ snap info edgex-device-modbus
```

The latest stable version of the snap can be installed using:

```bash
$ sudo snap install edgex-device-modbus
```

The 2.0 (Ireland) release of the snap can be installed using:

```bash
$ sudo snap install edgex-device-modbus --channel=2.0
```

The latest development version of the snap can be installed using:

```bash
$ sudo snap install edgex-device-modbus --edge
```

**Note** - the snap has only been tested on Ubuntu Core, Desktop, and Server.


## Snap configuration

Device services implement a service dependency check on startup which ensures that all of the runtime dependencies of a particular service are met before the service transitions to active state.

Snapd doesn't support orchestration between services in different snaps. It is therefore possible on a reboot for a device service to come up faster than all of the required services running in the main edgexfoundry snap. If this happens, it's possible that the device service repeatedly fails startup, and if it exceeds the systemd default limits, then it might be left in a failed state. This situation might be more likely on constrained hardware (e.g. RPi).

This snap therefore implements a basic retry loop with a maximum duration and sleep interval. If the dependent services are not available, the service sleeps for the defined interval (default: 1s) and then tries again up to a maximum duration (default: 60s). These values can be overridden with the following commands:
    
To change the maximum duration, use the following command:

```bash
$ sudo snap set edgex-device-modbus startup-duration=60
```

To change the interval between retries, use the following command:

```bash
$ sudo snap set edgex-device-modbus startup-interval=1
```

The service can then be started as follows. The "--enable" option
ensures that as well as starting the service now, it will be automatically started on boot:

```bash
$ sudo snap start --enable edgex-device-modbus.device-modbus
```

### Using a content interface to set device configuration

The `device-config` content interface allows another snap to seed this device
snap with both a configuration file and one or more device profiles.

To use, create a new snap with a directory containing the configuration and device 
profile files. Your snapcraft.yaml file then needs to define a slot with 
read access to the directory you are sharing.

```
slots:
  device-config:
    interface: content
    content: device-config
    read:
      - $SNAP/config
```

where `$SNAP/config` is configuration directory your snap is providing to 
the device snap.

Then connect the plug in the device snap to the slot in your snap,
which will replace the configuration in the device snap. Do this with:

```bash
$ sudo snap connect edgex-device-modbus:device-config your-snap:device-config
```

This needs to be done before the device service is started for the first time. 
Once you have set the configuration the device service can be started and 
it will then be configurated using the settings you provided:

```bash
$ sudo snap start edgex-device-modbus.device-modbus
```

**Note** - content interfaces from snaps installed from the Snap Store that have the same publisher connect automatically. For more information on snap content interfaces please refer to the snapcraft.io [Content Interface](https://snapcraft.io/docs/content-interface) documentation.

### Autostart
By default, the edgex-device-modbus snap disables its service on install, as the expectation is that the default profile configuration files will be customized, and thus this behavior allows the profile ```configuration.toml``` files in $SNAP_DATA to be modified before the service is first started.

This behavior can be overridden by setting the ```autostart``` configuration setting to "true". This is useful when configuration and/or device profiles are being provided via configuration or gadget snap content interface.

**Note** - this option is typically set from a gadget snap.

### Rich Configuration
While it's possible on Ubuntu Core to provide additional profiles via gadget
snap content interface, quite often only minor changes to existing profiles are required.

These changes can be accomplished via support for EdgeX environment variable
configuration overrides via the snap's configure and install hooks.
If the service has already been started, setting one of these overrides currently requires the
service to be restarted via the command-line or snapd's REST API.
If the overrides are provided via the snap configuration defaults capability of a gadget snap,
the overrides will be picked up when the services are first started.

The following syntax is used to specify service-specific configuration overrides:


```env.<stanza>.<config option>```

For instance, to setup an override of the service's Port use:

```$ sudo snap set edgex-device-modbus env.service.port=2112```

And restart the service:

```$ sudo snap restart edgex-device-modbus.device-modbus```

**Note** - at this time changes to configuration values in the [Writable] section are not supported.

For details on the mapping of configuration options to Config options, please refer to [Service Environment Configuration Overrides](#service-environment-configuration-overrides) below.

## Service Environment Configuration Overrides

**Note** - all of the configuration options below must be specified with the prefix: 'env.'

```
[Service]
service.health-check-interval   // Service.HealthCheckInterval
service.host                    // Service.Host
service.server-bind-addr        // Service.ServerBindAddr
service.port                    // Service.Port
service.max-result-count        // Service.MaxResultCount
service.max-request-size        // Service.MaxRequestSize
service.startup-msg             // Service.StartupMsg
service.request-timeout         // Service.RequestTimeout

[Clients.core-data]
clients.core-data.port          // Clients.core-data.Port

[Clients.core-metadata]
clients.core-metadata.port      // Clients.core-metadata.Port

[Device]
device.update-last-connected    // Device.UpdateLastConnected
device.use-message-bus          // Device.UseMessageBus
```
