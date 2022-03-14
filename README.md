# homestatus

Serves weather data (and maybe other information in the future) to a M5Paper device.

## M5paper

The code under m5paper requests data from the server, deserializes it, and renders to the screen. It shuts down (wakeup done by the real time clock) after it is finished rendering.

![image showing a m5paper rendering weather data](https://github.com/colonelxc/homestatus/blob/main/demo.jpg?raw=true)

Running the deserialization tests requires this single-file test library [https://github.com/colonelxc/cxxtest](https://github.com/colonelxc/cxxtest)


## Homestatus service

Service written in Go. It periodically loads weather data in the background, and serves the latest data upon request. It only listens on localhost, as I serve it behind Caddy.

### weather

Pulls weather forecast data from api.weather.gov

### Custom TSV-based protocol

The protocol is schemaless, but hopefully a bit easier to parse than json.

There can be one or more sections
1. A name for the section. This is essentially a hint to the parser for what is coming next.
2. A (tab separated) list of column names.
3. One or more data rows (with the same number of columns as above), with each data value separated by tabs.
4. An empty line (a newline following the last data row newlines)
