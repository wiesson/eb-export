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
eb-export -from 2017-08-01 -to 2017-09-26 -logger 5978f024ce9d0513a30000d7 -sensor 5978f024ce9d0513a300013c -sensor 5978f024ce9d0513a300013d -sensor 5978f024ce9d0513a300013e -sensor 598170a3ce9d053e4d0001e2 -aggr minutes_1 -type energy
```

#### -from*

example: `-from 2017-08-01`

#### -to*

example: `-to 2017-08-08`

#### -logger*

example: `-logger 12345678`

#### -sensor (array)*

example: `-sensor 23456789` or `-sensor 23456789 -sensor 34567890`

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
