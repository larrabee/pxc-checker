# Percona/MySQL XtraDB Cluster Checker
This project is [percona-clustercheck](https://github.com/olafz/percona-clustercheck) rewrited to Golang.
Program to make a proxy (ie HAProxy) capable of monitoring Percona XtraDB Cluster nodes properly.

## Usage

Basic Haproxy config
```
listen pxc
  bind 127.0.0.1:3306
  balance leastconn
  option httpchk
  mode tcp
  default-server check port 9200 inter 500 rise 5 fall 5
    server node1 1.2.3.4:3306 check port 9200
    server node2 1.2.3.5:3306 check port 9200
    server node3 1.2.3.6:3306 check port 9200 backup
```

