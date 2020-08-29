# wallmon

Small go cli for running wall monitors.

Many run grafana or other monitoring dashboards in chrome on a "forever on" screen. 
Chrome seems to have stability issues running for long times without restarts.
`wallmon` starts a fullscreen kiosk mode chrome with a specific webpage and periodically restarts it.

That's it.

usage:
```
$ ./wallmon -h
Usage of ./wallmon:
  -chrome-cmd string
        path to chrome cmd (default "/Applications/Google Chrome.app/Contents/MacOS/Google Chrome")
  -data-dir string
        the data-directory to use for chrome (default "/var/folders/vq/8r6mywnx7cvglrl9hvpyv4sh0000gn/T/wallmon-data-dir")
  -restart-delay duration
        delay between restarts of chrome (min 1s) (default 1s)
  -restart-interval duration
        restart interval of chrome (min 2s) (default 12h0m0s)
  -url string
        the uri to visit
```