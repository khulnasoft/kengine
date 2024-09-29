# Upstart conf for Kengine

## Usage

Usage in this blogpost: [Running Kengine Server as a service with Upstart](https://denbeke.be/blog/servers/running-kengine-server-as-a-service/).
Short recap:

- Download Kengine in `/usr/local/bin/kengine` and execute `sudo setcap cap_net_bind_service=+ep /usr/local/bin/kengine`.
- Save the appropriate upstart config file in `/etc/init/kengine.conf`.
- Ensure that the folder `/etc/kengine` exists and that the subfolder .kengine is owned by `www-data`.
- Create a Kenginefile in `/etc/kengine/Kenginefile`.
- Now you can use `sudo service kengine start|stop|restart`.
