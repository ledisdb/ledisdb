package server

import (
	"encoding/json"
	"github.com/siddontang/go-log/log"
	"io/ioutil"
	"os"
	"path"
)

type masterInfo struct {
	Addr    string `json:"addr"`
	LogFile string `json:"log_name"`
	LogPos  int64  `json:"log_pos"`
}

func (app *App) getMasterInfoName() string {
	return path.Join(app.cfg.DataDir, "master.info")
}

func (app *App) loadMasterInfo() error {
	data, err := ioutil.ReadFile(app.getMasterInfoName())
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		} else {
			return err
		}
	}

	if err = json.Unmarshal(data, &app.master); err != nil {
		return err
	}

	return nil
}

func (app *App) saveMasterInfo() error {
	bakName := path.Join(app.cfg.DataDir, "master.info.bak")

	data, err := json.Marshal(&app.master)
	if err != nil {
		return err
	}

	var fd *os.File
	fd, err = os.OpenFile(bakName, os.O_CREATE|os.O_WRONLY, os.ModePerm)
	if err != nil {
		return err
	}

	if _, err = fd.Write(data); err != nil {
		fd.Close()
		return err
	}

	fd.Close()
	return os.Rename(bakName, app.getMasterInfoName())
}

func (app *App) slaveof(masterAddr string) error {
	if len(masterAddr) == 0 {
		//stop replication
	} else {
	}

	return nil
}

func (app *App) runReplication() {
}

func (app *App) startReplication(masterAddr string) error {
	if err := app.loadMasterInfo(); err != nil {
		log.Error("load master.info error %s, use fullsync", err.Error())
		app.master = masterInfo{masterAddr, "", 0}
	} else if app.master.Addr != masterAddr {
		if err := app.ldb.FlushAll(); err != nil {
			log.Error("replication flush old data error %s", err.Error())
			return err
		}

		app.master = masterInfo{masterAddr, "", 0}
	}

	return nil
}
