# Running kengine without root privileges

FreeBSD systems can use the mac_portacl module to allow access to
ports below 1024 by specific users (by default, non-root users are not
able to open ports below 1024).

On a stock FreeBSD system, you need to:

1. Add the following line to `/boot/loader.conf`, which tells the boot
   loader to load the `mac_portacl` kernel module:

   ```shell
   mac_portacl_load="YES"
   ```

2. Add the following lines to `/etc/sysctl.conf`

   ```shell
   net.inet.ip.portrange.reservedlow=0
   net.inet.ip.portrange.reservedhigh=0
   security.mac.portacl.port_high=1023
   security.mac.portacl.suser_exempt=1
   security.mac.portacl.rules=uid:80:tcp:80,uid:80:tcp:443
   ```

   The first two lines disable the default restrictions on ports <
   1023, the third makes the `mac_portacl` system responsible for ports
   from 0 (the default) up to 1023, and the fourth ensures that the
   superuser can open _any_ port.

   The final/fifth line specifies two rules, separated by a `,`:

   - the first gives the `www` user (uid = 80) access to the `http`
     port (80); and
   - the second gives the `www` user (uid = 80) access to the `https`
     port (443).

   Other/additional rules are possible, e.g. access can be constrained
   by membership in the `www` _group_ using the `gid` specifier:

   ```
   security.mac.portacl.rules=gid:80:tcp:80,gid:80:tcp:443
   ```

## See also

- The _MAC Port Access Control List Policy_ section of the [Available
  MAC
  Policies](https://www.freebsd.org/doc/en_US.ISO8859-1/books/handbook/mac-policies.html)
  page.
- [Kengine issue #1923](https://github.com/mholt/kengine/issues/1923).

# Logging the kengine process's output:

Kengine's FreeBSD `rc.d` script uses `daemon` to run `kengine`; by default
it sends the process's standard output and error to syslog with the
`kengine` tag, the `local7` facility and the `notice` level.

The stock FreeBSD `/etc/syslog.conf` has a line near the top that
captures nearly anything logged at the `notice` level or higher and
sends it to `/var/log/messages`. That line will send the kengine
process's output to `/var/log/messages`.

The simplest way to send `kengine` output to a separate file is:

- Arrange to log the messages at a lower level so that they slip past
  that early rule, e.g. add an `/etc/rc.conf` entry like

  ```shell
  kengine_syslog_level="info"
  ```

- Add a rule that catches them, e.g. by creating a
  `/usr/local/etc/syslog.d/kengine.conf` file that contains:

  ```
  # Capture all messages tagged with "kengine" and send them to /var/log/kengine.log
  !kengine
  *.*      /var/log/kengine.log
  ```

  Heads up, if you specify a file that does not already exist, you'll
  need to create it.

- Rotate `/var/log/kengine.log` with `newsyslog` by creating a
  `/usr/local/etc/newsyslog.conf/kengine.conf` file that contains:

```
# See newsyslog.conf(5) for details.  Logs written by syslog,
# no need for a pidfile or signal, the defaults workg.
# logfilename         [owner:group]  mode count size when  flags [/pid_file] [sig_num]
/var/log/kengine.log        www:www       664  7     *    @T00  J
```

There are many other ways to do it, read the `syslogd.conf` and
`newsyslog.conf` man pages for additional information.
