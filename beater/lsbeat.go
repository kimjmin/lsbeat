package beater

import (
	"fmt"
	"io/ioutil"
	"time"

	"github.com/elastic/beats/libbeat/beat"
	"github.com/elastic/beats/libbeat/common"
	"github.com/elastic/beats/libbeat/logp"
	"github.com/elastic/beats/libbeat/publisher"

	"github.com/kimjmin/lsbeat/config"
)

type Lsbeat struct {
	done   chan struct{}
	config config.Config
	client publisher.Client

	lastIndexTime time.Time
}

// Creates beater
func New(b *beat.Beat, cfg *common.Config) (beat.Beater, error) {
	config := config.DefaultConfig
	if err := cfg.Unpack(&config); err != nil {
		return nil, fmt.Errorf("Error reading config file: %v", err)
	}

	ls := &Lsbeat{
		done:   make(chan struct{}),
		config: config,
	}
	return ls, nil
}

/// *** Beater interface methods ***///

func (bt *Lsbeat) Run(b *beat.Beat) error {
	logp.Info("lsbeat is running! Hit CTRL-C to stop it.")

	bt.client = b.Publisher.Connect()

	bt.listDir(bt.config.Path, b.Name, true) // init directory
	bt.lastIndexTime = time.Now()            // init indexing timestamp

	ticker := time.NewTicker(bt.config.Period)
	for {

		select {
		case <-bt.done:
			return nil
		case <-ticker.C:
		}

		now := time.Now()
		bt.listDir(bt.config.Path, b.Name, false) // call lsDir
		bt.lastIndexTime = now                    // mark Timestamp

		logp.Info("Event sent")
	}

}

func (bt *Lsbeat) Stop() {
	bt.client.Close()
	close(bt.done)
}

func (bt *Lsbeat) listDir(dirFile string, beatname string, init bool) {
	files, _ := ioutil.ReadDir(dirFile)
	for _, f := range files {
		t := f.ModTime()
		//fmt.Println(f.Name(), dirFile+"/"+f.Name(), f.IsDir(), t, f.Size())

		event := common.MapStr{
			"@timestamp": common.Time(time.Now()),
			"type":       beatname,
			"modtime":    common.Time(t),
			"filename":   f.Name(),
			"path":       dirFile + "/" + f.Name(),
			"directory":  f.IsDir(),
			"filesize":   f.Size(),
		}
		if init {
			// index all files and directories on init
			bt.client.PublishEvent(event) //elasticsearch index.
		} else {
			// Index only changed files since last run.
			if t.After(bt.lastIndexTime) {
				bt.client.PublishEvent(event) //elasticsearch index.
			}
		}

		if f.IsDir() {
			bt.listDir(dirFile+"/"+f.Name(), beatname, init)
		}
	}
}
