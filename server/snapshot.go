package server

import (
	"fmt"
	"github.com/siddontang/go/log"
	"github.com/siddontang/ledisdb/config"
	"io"
	"io/ioutil"
	"os"
	"path"
	"sort"
	"sync"
	"time"
)

const (
	snapshotTimeFormat = "2006-01-02T15:04:05"
)

type snapshotStore struct {
	sync.Mutex

	cfg *config.Config

	names []string

	quit chan struct{}
}

func snapshotName(t time.Time) string {
	return fmt.Sprintf("snap-%s.dump", t.Format(snapshotTimeFormat))
}

func parseSnapshotName(name string) (time.Time, error) {
	var timeString string
	if _, err := fmt.Sscanf(name, "snap-%s.dump", &timeString); err != nil {
		return time.Time{}, err
	}
	when, err := time.Parse(snapshotTimeFormat, timeString)
	if err != nil {
		return time.Time{}, err
	}
	return when, nil
}

func newSnapshotStore(cfg *config.Config) (*snapshotStore, error) {
	if len(cfg.Snapshot.Path) == 0 {
		cfg.Snapshot.Path = path.Join(cfg.DataDir, "snapshot")
	}

	if err := os.MkdirAll(cfg.Snapshot.Path, 0755); err != nil {
		return nil, err
	}

	s := new(snapshotStore)
	s.cfg = cfg
	s.names = make([]string, 0, s.cfg.Snapshot.MaxNum)

	s.quit = make(chan struct{})

	snapshots, err := ioutil.ReadDir(cfg.Snapshot.Path)
	if err != nil {
		return nil, err
	}

	for _, info := range snapshots {
		if _, err := parseSnapshotName(info.Name()); err != nil {
			log.Error("invalid snapshot file name %s, err: %s", info.Name(), err.Error())
			continue
		}

		s.names = append(s.names, info.Name())
	}

	//from old to new
	sort.Strings(s.names)

	go s.run()

	return s, nil
}

func (s *snapshotStore) Close() {
	close(s.quit)
}

func (s *snapshotStore) run() {
	t := time.NewTicker(1 * time.Minute)
	defer t.Stop()

	for {
		select {
		case <-t.C:
			s.purge()
		case <-s.quit:
			return
		}
	}
}

func (s *snapshotStore) purge() {
	s.Lock()
	var names []string
	num := s.cfg.Snapshot.MaxNum
	if len(s.names) > num {
		names = s.names[0:num]

		s.names = s.names[num:]
	}

	s.Unlock()

	for _, n := range names {
		if err := os.Remove(s.snapshotPath(n)); err != nil {
			log.Error("purge snapshot %s error %s", n, err.Error())
		}
	}
}

func (s *snapshotStore) snapshotPath(name string) string {
	return path.Join(s.cfg.Snapshot.Path, name)
}

type snapshotDumper interface {
	Dump(w io.Writer) error
}

func (s *snapshotStore) Create(d snapshotDumper) (*os.File, time.Time, error) {
	s.Lock()
	defer s.Unlock()

	now := time.Now()
	name := snapshotName(now)

	if len(s.names) > 0 && s.names[len(s.names)-1] >= name {
		return nil, time.Time{}, fmt.Errorf("create snapshot file time %s is behind %s ", name, s.names[len(s.names)-1])
	}

	f, err := os.OpenFile(s.snapshotPath(name), os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		return nil, time.Time{}, err
	}

	if err := d.Dump(f); err != nil {
		f.Close()
		os.Remove(s.snapshotPath(name))
		return nil, time.Time{}, err
	}

	s.names = append(s.names, name)

	f.Seek(0, os.SEEK_SET)

	return f, now, nil
}

func (s *snapshotStore) OpenLatest() (*os.File, time.Time, error) {
	s.Lock()
	defer s.Unlock()

	if len(s.names) == 0 {
		return nil, time.Time{}, nil
	}

	name := s.names[len(s.names)-1]
	t, _ := parseSnapshotName(name)

	f, err := os.Open(s.snapshotPath(name))
	if err != nil {
		return nil, time.Time{}, err
	}

	return f, t, err
}
