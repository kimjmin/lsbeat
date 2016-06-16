# Creating lsbeat

### install GoLang
```
wget https://storage.googleapis.com/golang/go1.6.2.linux-amd64.tar.gz
sudo tar -C /usr/local -xzf go1.6.2.linux-amd64.tar.gz
mkdir go
```

edit `.bash_profile`
```
export PATH=$PATH:/usr/local/go/bin
export GOPATH=$HOME/go
export PATH=$PATH:$GOPATH/bin
```

### install elastic stack
```
wget https://download.elastic.co/elasticsearch/release/org/elasticsearch/distribution/tar/elasticsearch/5.0.0-alpha3/elasticsearch-5.0.0-alpha3.tar.gz
wget https://download.elastic.co/kibana/kibana/kibana-5.0.0-alpha3-linux-x64.tar.gz
tar xvfz elasticsearch-5.0.0-alpha3.tar.gz; tar xvfz kibana-5.0.0-alpha3-linux-x64.tar.gz
```

### install JDK, git
```
sudo yum remove java -y
sudo yum install java-1.8.0-openjdk-devel.x86_64 -y
sudo yum install git -y
```

### install glide, cookiecutter

install glide
```
go get github.com/Masterminds/glide
```

install cookiecutter
```
sudo pip install cookiecutter
```

### lsscan : program lists files and directory.
get lsscan project
```
go get github.com/kimjmin/lsscan
```
This project code will be reused.
ioutil package doc : https://golang.org/pkg/io/ioutil/#example_ReadDir

### Create beat skeleton
install Beats package
```
mkdir $GOPATH/src/github.com/elastic
cd  $GOPATH/src/github.com/elastic
git clone https://github.com/elastic/beats.git
```
or
```
go get github.com/elastic/beats
```

run cookiecutter
```
cd $GOPATH/src/github.com/{user}
cookiecutter $GOPATH/src/github.com/elastic/beats/generate/beat
```
```
project_name [Examplebeat]: lsbeat
github_name [your-github-name]: {id}
beat [lsbeat]:
beat_path [github.com/kimjmin]:
full_name [Firstname Lastname]: {Full Name}
```

download dependencies
```
make init
make commit
```

view config file
```
vi lsbeat.yml
vi lsbeat.template.json
```

### configuration : add parameters
edit `etc/beat.yml`
```
vi etc/beat.yml
```
add `path` parameter.
```yml
lsbeat:
  # Defines how often an event is sent to the output
  period: 10s
  path: "."
```

check added parameters.
```
make update
vi lsbeat.yml
```

edit `config/config.go` so we can use new parameters.
```
vi config/config.go
```
```go
package config

type Config struct {
        Lsbeat LsbeatConfig
}

type LsbeatConfig struct {
        Period string `config:"period"`
        Path   string `config:"path"`
}
```

### check libbeat package
```
vi vendor/github.com/elastic/beats/libbeat/beat/beat.go
```
Beater interface
```go
type Beater interface {
        Config(*Beat) error  // Read and validate configuration.
        Setup(*Beat) error   // Initialize the Beat.
        Run(*Beat) error     // The main event loop. This method should block until signalled to stop by an invocation of the Stop() method.
        Cleanup(*Beat) error // Cleanup is invoked to perform any final clean-up prior to exiting.
        Stop()               // Stop is invoked to signal that the Run method should finish its execution. It will be invoked at most once.
}
```

### code beater/lsbeat.go
```
vi beater/lsbeat.go
```
add `io/ioutil` package
```go
import (
  "fmt"
  "time"
  "io/ioutil"
```

add `path`, `lastIndexTime` variable
```go
type Lsbeat struct {
  beatConfig *config.Config
  done       chan struct{}
  period     time.Duration
  client     publisher.Client
  path string
  lastIndexTime time.Time
}
```

add `path` init value
```go
func (bt *Lsbeat) Setup(b *beat.Beat) error {

  // Setting default period if not set
  if bt.beatConfig.Lsbeat.Period == "" {
    bt.beatConfig.Lsbeat.Period = "10s"
  }

        if bt.beatConfig.Lsbeat.Path == "" {
                bt.beatConfig.Lsbeat.Path = "."
        }
        bt.path = bt.beatConfig.Lsbeat.Path
```

add `listDir` function on the bottom
```go
func listDir(dirFile string, bt *Lsbeat, b *beat.Beat, counter int) {
	files, _ := ioutil.ReadDir(dirFile)
	for _, f := range files {
		t := f.ModTime()
		//fmt.Println(f.Name(), dirFile+"/"+f.Name(), f.IsDir(), t, f.Size())

		event := common.MapStr{
			"@timestamp":  common.Time(time.Now()),
			"type":        b.Name,
			"counter":     counter,
			"modTime":     t,
			"filename":    f.Name(),
			"fullname":    dirFile + "/" + f.Name(),
			"isDirectory": f.IsDir(),
			"fileSize":    f.Size(),
		}
		//index all files and directories for first routine.
		if counter == 1 {
			bt.client.PublishEvent(event) //elasticsearch index.
		} else {
			//after second routine, index only files and directories which created after previous routine
			if t.After(bt.lasIndexTime) {
				bt.client.PublishEvent(event) //elasticsearch index.
			}
		}

		if f.IsDir() {
			listDir(dirFile+"/"+f.Name(), bt, b, counter)
		}
	}
}
```




### install beats package
```
go get github.com/elastic/beats
```



lsbeat.template.json
```
{
  "mappings": {
    "_default_": {
      "_all": {
        "norms": false
      },
      "dynamic_templates": [
        {
          "fields": {
            "mapping": {
              "ignore_above": 1024,
              "type": "keyword"
            },
            "match_mapping_type": "string",
            "path_match": "fields.*"
          }
        }
      ],
      "properties": {
        "@timestamp": {
          "type": "date"
        },
        "beat": {
          "properties": {
            "hostname": {
              "ignore_above": 1024,
              "type": "keyword"
            },
            "name": {
              "ignore_above": 1024,
              "type": "keyword"
            }
          }
        },
         "counter": {
          "type": "integer"
        },
        "filename": {
          "norms": false,
          "type": "text",
          "analyzer": "ls_ngram_analyzer"
        },
        "fullname": {
          "ignore_above": 1024,
          "type": "keyword"
        },
        "modTime": {
          "type": "date"
        },
        "type": {
          "ignore_above": 1024,
          "type": "keyword"
        },
        "tags": {
          "ignore_above": 1024,
          "type": "keyword"
        }
      }
    }
  },
  "order": 0,
  "settings": {
    "index.refresh_interval": "5s",
    "analysis": {
      "analyzer": {
        "ls_ngram_analyzer": {
          "tokenizer": "ls_ngram_tokenizer"
        }
      },
      "tokenizer": {
        "ls_ngram_tokenizer": {
          "type": "ngram",
          "min_gram": "2",
          "token_chars": [
            "letter",
            "digit"
          ]
        }
      }
    }
  },
  "template": "lsbeat-*"
}
```




## metricbeat
metricbeat template
```
curl -XPUT 'http://localhost:9200/_template/metricbeat' -d@metricbeat.template.json
```
dashboard
```
../dev-tools/import_dashboards.sh -h
../dev-tools/import_dashboards.sh -d etc/kibana
```