## AISETL [![Build Status](https://travis-ci.org/eholzbach/aisetl.svg?branch=master)](https://travis-ci.org/eholzbach/aisetl)

AISETL reads udp packets containing AIS NMEA messages, extracts the data, and loads it into redis. It also provides a web server that displays data points on a map.

I'm feeding this with [rtl-ais](https://github.com/dgiardini/rtl-ais) using a RTL2832U usb sdr and a home made coaxial collinear antenna attached to a raspberry pi.

## Configuration

This uses [viper](https://github.com/spf13/viper) to resolve configuration files. JSON, TOML, YAML, and HCL are valid formats.

Example:
```yaml
listen: '0.0.0.0:10110'
redis: '127.0.0.1:6379'
forward:
  - '8.8.8.8:2100'
  - '8.8.7.7:7027'
```
