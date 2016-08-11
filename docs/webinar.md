# Creating lsbeat

Hi. I am Jongmin, working as Tech Evangelist at Elastic.
Today we will see about Beats and how we can create your own beat.

Here is products by Elastic. We have Elastic Stack the open source project. We have Kibana for UI, Elasticsearch for index and search engine for data, and Logstash and Beats for data ingesters.

Beats are lightweight shippers, which is different from Logstash.

Beat is very light program. It is written in GoLang and compiled as running binary file. Every beats have single purpose, so if you want collect specific data you have, and index to elasticsearch, you have to build your own Beats.

Currently, we have libbeat - the common library that enables you to create your own beat, Packagetbeat ...

Today we will start from very beginning. I will create new instance from AWS, and will install golang and will go over all configurations we need.

### install GoLang
```
wget https://storage.googleapis.com/golang/go1.6.2.linux-amd64.tar.gz
sudo tar -C /usr/local -xzf go1.6.2.linux-amd64.tar.gz
mkdir go
```

After installed golang, we have to set Goroot and gopath. Goroot is the location where golang program is located, and gopath is the location where all the go project we made and we downloaded for third party goes.

edit `.bash_profile`
```
export GOROOT=/usr/local/go/bin
export GOPATH=$HOME/go
export PATH=$PATH:$GOROOT/bin:$GOPATH/bin
```

We just finished setting with GoLang. Now lets download Elastic stack. We will install elasticsearch and kibana to index and monitor the data.
Lets download with 5. alpha. Kibana 5 has console within it, which we can use much easier.
### install elastic stack
```
wget https://download.elastic.co/elasticsearch/release/org/elasticsearch/distribution/tar/elasticsearch/5.0.0-alpha3/elasticsearch-5.0.0-alpha3.tar.gz
wget https://download.elastic.co/kibana/kibana/kibana-5.0.0-alpha3-linux-x64.tar.gz
tar xvfz elasticsearch-5.0.0-alpha3.tar.gz; tar xvfz kibana-5.0.0-alpha3-linux-x64.tar.gz
```
Now, we have to update Java. Currently installed 1.7 and we have to update to 1.8. Make shure you have to remove current version and install updated one.
### install JDK, git
```
sudo yum remove java -y
sudo yum install java-1.8.0-openjdk-devel.x86_64 -y
sudo yum install git -y
```

And now, lets look the program code we are going to build. We are going to build lsbeat, which indexes list of files and directories to elasticsearch. When you run ls command at the unix based system, it will list up files and directories and some informations of it. We will going to make this kinds of program through golang. And code is already in the blog post.

This code starts with main function, and will check the argument. If there is no aregument, it will put dot, which means current path. And main function calls listDir function, which looks up files and directories and prints it's name, path, is directory, and size.
And if files is directory, it will run recursion of listDir function again, and print it's inner files list.

lets copy the code,
and make file, names gols.go
and lets run the command, go run gols.go

### install glide, cookiecutter

To use beat generator, we need two programs, glide and cookiecutter. We can install glide with go get command. it will download glide from github repository and will compile to bianry so we can use.

install glide
```
go get github.com/Masterminds/glide
```

And we can install cookiecutter throug pip.

install cookiecutter
```
sudo pip install cookiecutter
```

### Create beat skeleton
install Beats package. Like glide, lets use go get command to get elastic/beats repository from github.
```
go get github.com/elastic/beats
```

now, we have to create our path under github, which is kimjmin for me, 
and lets move to there and run cookiecutter to create skeleton.

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

and lets run make setup to download dependency libraries.
```
make setup
```

And here is the command to push our lsbeat project to remote github repository, but I will not going to run this, because I already have this project on my github account.

Now, let see configuration. Basically beat generator creates lsbeat.yml file with default settings in it which all beats contains. 
period is the parameter that descides interval of routine of Beats. Like here, lsbeat will repeat it's data harvesting logic every 1 second.

To change configuration, we have to edit beat.yml under etc directory.
Lets change it to 10.
And we will add new parameter, path, which represents the root path we will collect list.

and we reun make update command to apply new settings to lsbeat.yml file.


And we have to change config/config.go file, so our beat program can use the settings we added.


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

add `listDir()` function on the bottom
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
			if t.After(bt.lastIndexTime) {
				bt.client.PublishEvent(event) //elasticsearch index.
			}
		}

		if f.IsDir() {
			listDir(dirFile+"/"+f.Name(), bt, b, counter)
		}
	}
}
```

edit `Run()`, function to call `listDir()`
```go
func (bt *Lsbeat) Run(b *beat.Beat) error {
	logp.Info("lsbeat is running! Hit CTRL-C to stop it.")

	ticker := time.NewTicker(bt.period)
	counter := 1
	for {
		listDir(bt.path, bt, b, counter) // call lsDir
		bt.lastIndexTime = time.Now()     // mark Timestamp

		select {
		case <-bt.done:
			return nil
		case <-ticker.C:
		}

		// Remove previous logic.
		// event := common.MapStr{
		// 	"@timestamp": common.Time(time.Now()),
		// 	"type":       b.Name,
		// 	"counter":    counter,
		// }
		// bt.client.PublishEvent(event)

		logp.Info("Event sent")
		counter++
	}
}
```

### add new field informations
open `etc/fields.yml`
```
vi etc/fields.yml
```
add fields
```yml
- key: lsbeat
  title: LS Beat
  descrtiption: 
  fields:
    - name: counter
      type: integer
      required: true
      description: >
        PLEASE UPDATE DOCUMENTATION
    #new fiels added lsbeat
    - name: modTime
      type: date
    - name: filename
      type: text
    - name: fullname
    - name: isDirectory
      type: boolean
    - name: fileSize
      type: long
```

apply new updates
```
make update
```

set mappings and analyzer `lsbeat.template.json`
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

### build and run

```
make
```
modify `lsbeat.yml` file for scanning root directory

run elasticsearch and kibana
```
./lsbeat
```
