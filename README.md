# dm200stats

Simple daemon that polls a [Netgear DM200](http://www.netgear.com.au/home/products/networking/modem-routers/DM200.aspx) modem for basic connection information and provides an HTTP export suitable for collection by [Prometheus](http://prometheus.io).

## Usage

```shell
$ go build github.com/dparrish/dm200stats
$ ./dm200stats -user <username> -pass <password> -port 8080 10.0.0.1
```

Statistics will be exported by HTTP on the supplied `-port`.
