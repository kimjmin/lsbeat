# Create your own beat

Beats is the platform for building lightweight, open source data shippers for many types of data for Elasticsearch and Logstash. We are providing kinds of beats like Libbeat, Packetbeat, Filebeat and Metricbeat and also there are [Community Beats](https://www.elastic.co/guide/en/beats/libbeat/current/community-beats.html) made by beats community.

We provide a libbeat library and Beat Generator package that helps you to create your own beat. Today we will see how to create you own beat by using Beat Generator. The Beat that we create today for practice is `lsbeat`. lsbeat indexes informations of files and directories that lists with Unix command `ls`. This article is based with Unix OS, so if you are Windows or other OS user, follow informations which fits with your OS.


## Step 1 - Setup your Golang Environment

Beats are written in Golang. To create and develop a beat, Golang must be installed on your machine. Follow the guide here to [install Golang](https://golang.org/doc/install). Currently Beats require at least Golang 1.6. Make sure properly setup your `$GOPATH` variable. 

In addition to Golang Glide is used for the dependency management. Make sure to install at least Glide 0.10.0 from this guide [here](https://github.com/Masterminds/glide). We will need glide later.

By the way, let's see the code that we will use for lsbeat. This is simple golang program that lists all directories and files under itself and subdirectories of its runtime argument.

```go
package main

import (
	"fmt"
	"io/ioutil"
	"os"
)

func main() {
	//apply run path "." without argument.
	if len(os.Args) == 1 {
		listDir(".")
	} else {
		listDir(os.Args[1])
	}
}

func listDir(dirFile string) {
	files, _ := ioutil.ReadDir(dirFile)
	for _, f := range files {
		t := f.ModTime()
		fmt.Println(f.Name(), dirFile+"/"+f.Name(), f.IsDir(), t, f.Size())
    
		if f.IsDir() {
			listDir(dirFile + "/" + f.Name())
		}
	}
}
```

We will reuse the code of `listDir` function.

## Step 2 - Generate

To generate your own beat we use the beat-generator. First you must install `cookiecutter`. Check out the installations guides [here](http://cookiecutter.readthedocs.org/en/latest/installation.html). After having installed cookiecutter, we must decide for a name of our beat. The name must be one word all lower case. In our example which is `lsbeat`.

To create beat skeleton, you should download Beats Generator package, which included in `github.com/elastic/beats` repository. Once you installed GoLang, you can download it with using command `go get`. Once you run the command, all source files will be downloaded under `$GOPATH/src` path.
```
go get github.com/elastic/beats
```

Now created and move to your own repository under GOPATH, and run `cookiecutter` with Beat Generator path.
```
cd $GOPATH/src/github.com/{user}
cookiecutter $GOPATH/src/github.com/elastic/beats/generate/beat
```
Cookiecutter will ask you several questions. For your project_name enter `lsbeat`, for github_user - `your github id`. The next two question with for beat and beat_path should already be automatically set correct. For the last one your can insert your `Firstname Lastname`.
```
project_name [Examplebeat]: lsbeat
github_name [your-github-name]: {user}
beat [lsbeat]:
beat_path [github.com/{github id}]:
full_name [Firstname Lastname]: {Full Name}
```

This should now have created a directory `lsbeat` inside our folder with several files. Lets change to this directory and list up files automatically created.

```sh
$ cd lsbeat
$ tree
.
├── CONTRIBUTING.md
├── LICENSE
├── Makefile
├── README.md
├── beater
│   └── lsbeat.go
├── config
│   ├── config.go
│   └── config_test.go
├── dev-tools
│   └── packer
│       ├── Makefile
│       ├── beats
│       │   └── lsbeat.yml
│       └── version.yml
├── docs
│   └── index.asciidoc
├── etc
│   ├── beat.yml
│   └── fields.yml
├── glide.yaml
├── lsbeat.template.json
├── main.go
├── main_test.go
└── tests
    └── system
        ├── config
        │   └── lsbeat.yml.j2
        ├── lsbeat.py
        ├── requirements.txt
        └── test_base.py
```

We now have a raw template of our beat but still need to fetch the dependencies and setup our git repository.

```
make init
```
This command fetches the dependencies, which in our case is libbeat, creates the basic config and template files and starts a new git repository. We will have a closer look at the template and config files later. Make init can take some time to download libbeat.

After init is completed, we create the first commits of our repository with:

```
make commit
```

It will create a clean git history for each major step. Note that you can always rewrite the history if you wish before pushing your changes. From now on you can use all the normal git commands you are used to. To push `lsbeat` in the git repository, run the following commands:

```
git remote set-url origin https://github.com/{user}/lsbeat
git push origin master
```

Now we have a complete beat and pushed the first to Github. Lets build and run our beat and then dig deeper into the code.

## Step 3 - Set Configurations

Once you run commands above, it will create `lsbeat.yml`, `lsbeat.template.json` config files automatically. All basic configurations are already written in these files.

> lsbeat.yml

```yml
lsbeat:
  # Defines how often an event is sent to the output
  period: 1s
```

`period` is common parameter that included in all kind of beats. It represents that lsbeat iterates the process every 1 second. Let's change this period from 1 to 10 sec and add new `path` parameter which represents the path of top directory program will scan. We can add parameter in `beat.yml` under `etc` directory. Once we added new parameters, we run `make update` command to apply changes to config files.

> edit `etc/beat.yml`
```
vi etc/beat.yml
```
> add `path` parameter.
```yml
lsbeat:
  # Defines how often an event is sent to the output
  period: 10s
  path: "."
```

Lets run `make update` and check `lsbeat.yml` again. We can see the new parameter we set on `etc/beat.yml` is applied on lsbeat.yml
```sh
$ make update
$ cat lsbeat.yml
################### Lsbeat Configuration Example #########################

############################# Lsbeat ######################################

lsbeat:
  # Defines how often an event is sent to the output
  period: 10s
  path: "."
###############################################################################
```

After changing yml files, you should edit `config/config.go` so we can use new parameters.

> edit config/config.go

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
