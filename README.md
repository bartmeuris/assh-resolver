[![Build Status](https://travis-ci.org/bartmeuris/assh-resolver.svg?branch=master)](https://travis-ci.org/bartmeuris/assh-resolver)

# Resolve command for ssh

I use this in combination with [assh](https://github.com/moul/advanced-ssh-config) as the `ResolveCommand` to use different IP's to connect to hosts, depending on the location I am to speed up connecting to a host when a possible jump-host must be used.

Location detection is done by checking the default gateway, for which an external package is used. In the future - possible other ways of detecting the location might be implemented as required.

## Configuration

The configuration file is looked for here, in the following order:

* The file set in the `ASSH_RESOLVECFG`
* `.ssh/locations.yml` in the user homedir
* Only in debug builds: `locations.yml` in current directory

The format of the file is pretty simple:

```
---
test:
    # No gateway defined, this will be treated as default location
    short: tst

client1:
    short: cl1
    gateway: 10.0.0.1

client2:
    short: cl2
    gateway: 172.16.0.1

home:
    short: h
    gateway: 192.168.50.254

```

## Configuring ASSH

As resolve command you then use the following in the assh config:

    defaults:
      ResolveCommand: '/path/to/assh-resolver "%h"'

As hosts, you then specify all IP/hostnames linked with the location name in the following format:

    ([location|short];)hostname|...

The last entry without a location defined will be treated as the fallback if the current location could not be detected, or if it was not found in the host list.

Example:

    cl1;10.0.0.20|home;host.vpn.cl1|public.host.ip:2222

With the above configuration file example, this will:

* use `10.0.0.20` when the `client1` location is detected
* use `host.vpn.cl1` when the `home` location is detected
* if the location was neither `home` or `client1`, `public.host.ip:2222` is used.

You can also do this to specify jump-hosts in assh:

    cl1;10.0.0.20|cl1jumphost/10.0.0.20

This would:

* use `10.0.0.20` when the `client1` location is detected
* All other locations would use the "cl1jumphost" as gateway to connect to 10.0.0.20

This speeds up the connection speed so assh doesn't have to attempt methods that won't work, or would be inefficient.

