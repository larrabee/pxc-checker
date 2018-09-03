# Percona/MySQL XtraDB Cluster Checker
This project is [percona-clustercheck](https://github.com/olafz/percona-clustercheck) like checker rewrited to Golang. 
Program to make a proxy (ie HAProxy) capable of monitoring Percona XtraDB Cluster nodes properly.

## Usage

Basic Haproxy config:
```
listen pxc
  bind 127.0.0.1:3306
  balance leastconn
  option httpchk HEAD /
  mode tcp
  default-server inter 500 rise 5 fall 5
    server node1 1.2.3.4:3306 check port 9200
    server node2 1.2.3.5:3306 check port 9200
    server node3 1.2.3.6:3306 check port 9200 backup
```

## Setup

1. Create MySQL user:
    ```sql
    create user 'pxc_checker'@'localhost' IDENTIFIED BY 'YourStrongPassword'; GRANT PROCESS ON *.* TO 'pxc_checker'@'localhost';
    ```
2. Get program binary. You can choose one of the following methods:
    -  Build it from source code with:
          ```
          go get
          go build -o pxc-checker ./...
          ```
    - Download latest compiled binary from [Releases page](https://github.com/larrabee/pxc-checker/releases).
3. Copy binary to `/usr/bin/pxc-checker`
4. Copy systemd unit from `systemd/pxc-check@.service` to `/etc/systemd/system/pxc-checker@.service`
5. Copy example config from `config/example.conf` to `/etc/pxc/checker/cluster.conf` and modify it.
6. Enable and start unit with command: `systemctl enable --now pxc-checker@cluster`
7. Check node status with command: `curl http://127.0.0.1:9200`

## Configuration file options
You can override any of the following values in configuration file:

- `WEB_LISTEN`: Web server listening interface and port in format `{IPADDR}:{PORT}` or `:PORT` for all interfaces. Default: `:9200`
- `WEB_READ_TIMEOUT`: Web server request read timeout in milliseconds. Default: `30000`
- `WEB_WRITE_TIMEOUT`: Web server request write timeout in milliseconds. Default: `30000`
- `CHECK_RO_ENABLED`: Mark 'read_only' node as available. Default: `false`
- `CHECK_FORCE_ENABLE`: Ignoring the status of the checks and always marking the node as available. Default: `false`
- `CHECK_INTERVAL`: Mysql checks interval in milliseconds. Default: `500`
- `CHECK_FAIL_TIMEOUT`: Mark the node inaccessible if for the specified time (in milliseconds) there were no successful checks. Default: `3000`
- `MYSQL_HOST`: MySQL host address. Default: `127.0.0.1`
- `MYSQL_PORT`: MySQL port. Default: `3306`
- `MYSQL_USER`: MySQL username. Default: `pxc_checker`
- `MYSQL_PASS`: Mysql password. Default: no password

