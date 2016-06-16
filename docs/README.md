# Creating lsbeat

### install GoLang
```
wget https://storage.googleapis.com/golang/go1.6.2.linux-amd64.tar.gz
sudo tar -C /usr/local -xzf go1.6.2.linux-amd64.tar.gz
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
```

### install JDK, git
```
sudo yum install java-1.8.0-openjdk-devel.x86_64
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

https://golang.org/pkg/io/ioutil/#example_ReadDir

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