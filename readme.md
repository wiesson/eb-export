## API CSV Exporter

Commandline application to export large chunks of CSV values from the API

### Install

```
go get github.com/wiesson/eb-export
cd $GOPATH/src/github.com/wiesson/eb-export
go install
```

#### Binaries

Coming soon

### arguments

Example

```
eb-export -token <your-access-token> -from 2017-08-01 -to 2017-09-26 -logger 01234567 -sensor 12345678 -sensor 23456789 -sensor 34567890 -sensor 45678901 -aggr minutes_1 -type energy
```

#### -token\*

example: `-token <your-access-token>`

#### -from\*

example: `-from 2017-08-01`

#### -to\*

example: `-to 2017-08-08`

#### -logger\*

example: `-logger 12345678`

#### -sensor (array)\*

example: `-sensor 12345678` or `-sensor 12345678 -sensor 23456789`

#### -aggr

example: `-aggr days_1`
default `minutes_1`
options `days_1`, `hours_1`, `minutes_15`, `minutes_1` or `none` (probably slow)

#### -type

example: `-aggr days_1`
default `power`
options `power` or `energy`

#### -tz

example: `-tz Europe/Berlin`
default `UTC`
