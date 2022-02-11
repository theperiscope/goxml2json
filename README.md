# goxml2json

Simple tool to facilitate migration from .NET's _web.config_ or _app.config_ files' `appSettings` and `connectionStrings` sections to the JSON format used by the Azure App Services configuration. Supports also the `configSource` attribute to help read from connected configuration files.

## Usage

### Convert AppSettings and Designate Slot Settings

```
goxml2json appSettings -i Web.config -o appSettings.json -slotSetting setting1 -slotSetting settings
```

### Convert ConnectionStrings

```
goxml2json connectionStrings -i Web.config -o connectionStrings.json
```
