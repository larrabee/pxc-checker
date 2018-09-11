package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"github.com/labstack/gommon/log"
	"github.com/valyala/fasthttp"
	"strconv"
	"time"
)

type ReasonCode int

const (
	reasonInternalError     ReasonCode = -1
	reasonOk                ReasonCode = 0
	reasonForceEnabled      ReasonCode = 1
	reasonNodeNotAvailable  ReasonCode = 2
	reasonWSRepFailed       ReasonCode = 3
	reasonCheckTimeout      ReasonCode = 4
	reasonRWDisabled        ReasonCode = 5
	reasonNonPrimaryCluster ReasonCode = 6
)

type Response struct {
	*NodeStatus
	ReasonText string
	ReasonCode ReasonCode
}

func checkerHandler(ctx *fasthttp.RequestCtx) {
	response := Response{NodeStatus: status}
	ctx.SetContentType("application/json")

	if config.CheckForceEnable {
		ctx.SetStatusCode(fasthttp.StatusOK)
		response.ReasonText = "Force enabled"
		response.ReasonCode = reasonForceEnabled
	} else if !status.NodeAvailable {
		ctx.SetStatusCode(fasthttp.StatusServiceUnavailable)
		response.ReasonText = "Node is not available"
		response.ReasonCode = reasonNodeNotAvailable
	} else if !status.ClusterPrimary {
		ctx.SetStatusCode(fasthttp.StatusServiceUnavailable)
		response.ReasonText = "Node in non-Primary cluster"
		response.ReasonCode = reasonNonPrimaryCluster
	} else if (status.WSRepStatus != 4) && (status.WSRepStatus != 2) {
		ctx.SetStatusCode(fasthttp.StatusServiceUnavailable)
		response.ReasonText = "WSRep failed"
		response.ReasonCode = reasonWSRepFailed
	} else if status.Timestamp+config.CheckFailTimeout < unixTimestampMillisecond() {
		ctx.SetStatusCode(fasthttp.StatusServiceUnavailable)
		response.ReasonText = "Check timeout"
		response.ReasonCode = reasonCheckTimeout
	} else if !config.CheckROEnabled && !status.RWEnabled {
		ctx.SetStatusCode(fasthttp.StatusServiceUnavailable)
		response.ReasonText = "Node is read only"
		response.ReasonCode = reasonRWDisabled
	} else {
		ctx.SetStatusCode(fasthttp.StatusOK)
		response.ReasonText = "OK"
		response.ReasonCode = reasonOk
	}

	if ctx.IsGet() {
		if respJson, err := json.Marshal(response); err != nil {
			errStr := fmt.Sprintf(`{"ReasonText":"Internal checker error","ReasonCode":%d,"err":"%s"}`, reasonInternalError, err)
			ctx.SetBody([]byte(errStr))
			ctx.SetStatusCode(fasthttp.StatusInternalServerError)
		} else {
			ctx.SetBody(respJson)
		}
	}

	return
}

func checker(status *NodeStatus) {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/", config.MysqlUser, config.MysqlPass, config.MysqlHost, config.MysqlPort)
	log.Printf("Connecting to mysql with dsn: %s", dsn)
	dbConn, _ := sql.Open("mysql", dsn)

	for {
		sleepRemain(status.Timestamp, config.CheckInterval)
		curStatus := &NodeStatus{}
		curStatus.Timestamp = unixTimestampMillisecond()
		var key, value string

		row := dbConn.QueryRow("SHOW GLOBAL VARIABLES LIKE 'read_only';")

		err := row.Scan(&key, &value)
		if err != nil {
			goto Finish
		}
		if value == "OFF" {
			curStatus.RWEnabled = true
		}

		curStatus.NodeAvailable = true

		row = dbConn.QueryRow("SHOW STATUS LIKE 'wsrep_local_state';")
		err = row.Scan(&key, &value)
		if err != nil {
			goto Finish
		}
		curStatus.WSRepStatus, _ = strconv.Atoi(value)

		row = dbConn.QueryRow("SHOW STATUS LIKE 'wsrep_cluster_status';")
		err = row.Scan(&key, &value)
		if err != nil {
			goto Finish
		}
		if value == "Primary" {
			curStatus.ClusterPrimary = true
		}

	Finish:
		*status = *curStatus
	}
}

func sleepRemain(startTime int64, sleepTime int64) {
	curTime := unixTimestampMillisecond()
	actualSleepTime := sleepTime - (curTime - startTime)

	if actualSleepTime <= 0 {
		return
	} else if actualSleepTime >= sleepTime {
		time.Sleep(time.Duration(sleepTime) * time.Millisecond)
	} else {
		time.Sleep(time.Duration(actualSleepTime) * time.Millisecond)
	}

	return
}

func unixTimestampMillisecond() int64 {
	return time.Now().UnixNano() / (int64(time.Millisecond) / int64(time.Nanosecond))
}
