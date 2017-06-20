# Isowrap

Isowrap is a library used to execute programs isolated from the rest of the system.

It is a wrapper around Linux Containers (using [isolate](https://github.com/ioi/isolate)) and FreeBSD [jails](https://www.freebsd.org/doc/handbook/jails.html) (WIP).

This is probably alpha quality software.

## To do:

- [x] Linux isolate runner
  - [ ] Full env
- [x] FreeBSD jail runner
  - [x] Implement "proper" wall time limit.
  - [x] Stack limit
  - [x] Maximum number of processes
  - [ ] Enable/Disable networking
  - [x] Environment variables

## Platform specific requirements

### Linux (`isolate`)

See the [INSTALLATION](https://github.com/ioi/isolate/blob/master/isolate.1.txt#L254-L280) part of the isolate manual. Control groups are required, make sure that they are enabled and `cgroupfs` is mounted.

### FreeBSD (`jail`)

Enable kernel `racct` support by adding the following line to `/etc/loader.conf`:

```
kern.racct.enable=1
```
