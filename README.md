# dumpils

dumpils demodulates ILS signal and dumps the measurements
in various formats.

Initial version without channel filter.

## Running

dumpils reads measurements from a rtl_tcp server.

Currently dumpils connects to a local rtl_tcp server
(listening at localhost:1234).

The ILS channel is hardcoded at 110.1 MHz.

Run with ```go run cmd/dumpils/main.go```

## Example
![dumpils screenshot](example.png?raw=true)
