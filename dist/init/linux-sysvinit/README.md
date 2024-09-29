# SysVinit conf for Kengine

## Usage

- Download the appropriate Kengine binary in `/usr/local/bin/kengine` or use `curl https://getkengine.com | bash`.
- Save the SysVinit config file in `/etc/init.d/kengine`.
- Ensure that the folder `/etc/kengine` exists and that the folder `/etc/ssl/kengine` is owned by `www-data`.
- Create a Kenginefile in `/etc/kengine/Kenginefile`
- Now you can use `service kengine start|stop|restart|reload|status` as `root`.

## Init script manipulation

The init script supports configuration via the following files:

- `/etc/default/kengine` ( Debian based https://www.debian.org/doc/manuals/debian-reference/ch03.en.html#_the_default_parameter_for_each_init_script )
- `/etc/sysconfig/kengine` ( CentOS based https://www.centos.org/docs/5/html/5.2/Deployment_Guide/s1-sysconfig-files.html )

The following variables can be changed:

- DAEMON: path to the kengine binary file (default: `/usr/local/bin/kengine`)
- DAEMONUSER: user used to run kengine (default: `www-data`)
- PIDFILE: path to the pidfile (default: `/var/run/$NAME.pid`)
- LOGFILE: path to the log file for kengine daemon (not for access logs) (default: `/var/log/$NAME.log`)
- CONFIGFILE: path to the kengine configuration file (default: `/etc/kengine/Kenginefile`)
- KENGINEPATH: path for SSL certificates managed by kengine (default: `/etc/ssl/kengine`)
- ULIMIT: open files limit (default: `8192`)
