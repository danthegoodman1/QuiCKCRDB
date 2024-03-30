# QuiCKCRDB

The Apple QuiCK queue implemented on top of CockroachDB

## Logging

QuiCKCRDB uses zerolog in JSON format, output to stdout

`QUICK_DEBUG=1` enables debug logging
`QUICK_INFO=1` enables info logging
`QUICK_DISABLE_LOG=1` disables all logging (not advised)

Otherwise, the log level is `warn`.

## Schema Setup

You are required to create the tables found in `tables.sql`, with the provided names.