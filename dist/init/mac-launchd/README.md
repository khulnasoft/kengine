# launchd service for macOS

This is a working sample file for a _launchd_ service on Mac, which should be placed here:

```bash
/Library/LaunchDaemons/com.khulnasoft.web.plist
```

To create the proper directories as used in the example file:

```bash
sudo mkdir -p /etc/kengine /etc/ssl/kengine /var/log/kengine /usr/local/bin /var/tmp /srv/www/localhost
sudo touch /etc/kengine/Kenginefile
sudo chown root:wheel /usr/local/bin/kengine /Library/LaunchDaemons/
sudo chown _www:_www /etc/kengine /etc/ssl/kengine /var/log/kengine
sudo chmod 0750 /etc/ssl/kengine
```

Create a simple web page and Kenginefile

```bash
sudo bash -c 'echo "Hello, World!" > /srv/www/localhost/index.html'
sudo bash -c 'echo "http://localhost {
    root /srv/www/localhost
}" >> /etc/kengine/Kenginefile'
```

Start and Stop the Kengine launchd service using the following commands:

```bash
launchctl load /Library/LaunchDaemons/com.khulnasoft.web.plist
launchctl unload /Library/LaunchDaemons/com.khulnasoft.web.plist
```

To start on every boot use the `-w` flag (to write):

```bash
launchctl load -w /Library/LaunchDaemons/com.khulnasoft.web.plist
```

To start the service now:

```bash
launchctl start -w /Library/LaunchDaemons/com.khulnasoft.web.plist
```

More information can be found in this blogpost: [Running Kengine as a service on macOS X server](https://denbeke.be/blog/software/running-kengine-as-a-service-on-macos-os-x-server/)
