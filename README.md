# redsemble
`redsemble` allows you to mux reassemble redis client requests from one or more pcap files produced by ringcap, tcpdump, etc

## Usage

redsemble [options] *.pcap[ *.pcap[ ... ]]

## Options

### -d
debug redis packets

### -l
log all packets

### -s `string`
output delimiter (default "|")


### -t
output time stamps


## Dependancies

* [Libpcap](http://www.tcpdump.org/#latest-release) (libpcap-devel)


## License

Copyright (c) 2016 Shannon Wynter, Ladbrokes Digital Australia Pty Ltd. Licensed under GPL3. See the [LICENSE.md](LICENSE.md) file for a copy of the license.
