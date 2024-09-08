# Network RX Tracker

Network RX Tracker is a tool for monitoring network ingress statistics using eBPF. It provides real-time insights into network traffic on a specified network interface.

## Features

- Real-time monitoring of network ingress statistics
- Multiple display modes: text, TUI (Terminal User Interface), and aggregate
- Configurable refresh interval and time window
- Logging support
- Ability to save charts

## Installation

For x86 architecture, you can download the pre-built binary from here [binary](https://github.com/raghu-nandan-bs/nw-rx-tracker/blob/main/release/nw-rx-tracker).

For other architectures, you can build the binary yourself using the following command:

```
sudo make release
```

## Usage

Run the Network RX Tracker with the following command:

```
sudo ./nw-rx-tracker [flags]
```

### TUI mode

`sudo ./release/nw-rx-tracker --device enp0s31f6 --display tui --interval 100ms`

![TUI mode](./docs/tui.png)

### Text mode 

`sudo ./release/nw-rx-tracker --device enp0s31f6 --display text --interval 1s`
example output
```
--- Inbound Network Statistics at 2024-09-08 14:38:35.848975 ---

Inbound IPv4 Traffic:
  35.174.210.7                             Bytes: 160 B       Packets: 1

Inbound IPv6 Traffic:
  400::4100:2028:816:2:524                 Bytes: 5.5 MB      Packets: 3647
  1000::2517:816:2:524                     Bytes: 280 kB      Packets: 220

--- Inbound Network Statistics at 2024-09-08 14:38:37.849243 ---

Inbound IPv4 Traffic:
  23.41.186.66                             Bytes: 888 B       Packets: 7
  35.174.210.7                             Bytes: 164 B       Packets: 1
  54.173.95.250                            Bytes: 164 B       Packets: 1
  192.168.29.1                             Bytes: 60 B        Packets: 1

Inbound IPv6 Traffic:
  e20::1508:240:68:424                     Bytes: 563 B       Packets: 6
  11d:a8c0::db28:35d0:102:524              Bytes: 587 B       Packets: 5
  400::4100:2028:816:2:524                 Bytes: 5.6 MB      Packets: 3701

--- Inbound Network Statistics at 2024-09-08 14:38:39.848503 ---

Inbound IPv4 Traffic:
  192.168.29.1                             Bytes: 60 B        Packets: 1
  3.225.222.30                             Bytes: 160 B       Packets: 1

Inbound IPv6 Traffic:
  56a5:defe:ff99:a38e::80fe                Bytes: 90 B        Packets: 1
  400::4100:2028:816:2:524                 Bytes: 93 kB       Packets: 1013

--- Inbound Network Statistics at 2024-09-08 14:38:41.849079 ---

Inbound IPv4 Traffic:
  54.173.95.250                            Bytes: 160 B       Packets: 1

Inbound IPv6 Traffic:
  1000::2517:816:2:524                     Bytes: 266 kB      Packets: 209
  400::4100:2028:816:2:524                 Bytes: 26 kB       Packets: 297
```


### Flags

- `--device`: Network device name (default: "eth0")
- `--interval`: Interval for refreshing stats (default: "1s")
- `--window`: Width of the TUI time series (default: "30s")
- `--display`: Display mode (options: text, tui, aggregate) (default: "plain")
- `--log`: Log file path (default: "/tmp/nwrxtrkr.log")
- `--log-level`: Log level (options: trace, debug, info, warn, error, fatal, panic) (default: "info")
- `--save`: Path to save the chart (default: "/tmp/nwrxtrkr.html")

### Examples

1. Monitor wlan0 with default settings:
   ```
   sudo ./nw-rx-tracker -device wlan0
   ```

2. Monitor eth0 with a 5-second refresh interval and TUI display:
   ```
   sudo ./nw-rx-tracker -device eth0 -interval 5s -display tui
   ```

3. Monitor eth1 with a 1-minute window and save the chart:
   ```
   sudo ./nw-rx-tracker -device eth1 -window 1m -save ~/network_chart.html
   ```

## Logging

Logs are written to the specified log file (default: `/tmp/nwrxtrkr.log`). You can change the log level using the `--log-level` flag.

## Saving Charts

You can save a chart of the network statistics by specifying the `--save` flag with a file path.

## REFERENCES

https://ebpf-go.dev/

