# homestatus

Serves weather information to a M5Paper device.

## weather

Pulls weather forecast data from api.weather.gov

## Custom TSV-based protocol

The protocol is schemaless, but hopefully a bit easier to parse than json.

There can be one or more sections
1. A name for the section. This is essentially a hint to the parser for what is coming next.
2. A (tab separated) list of column names.
3. One or more data rows (with the same number of columns as above), with each data value separated by tabs.
4. An empty line (a newline following the last data row newlines)
