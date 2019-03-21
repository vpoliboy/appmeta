## Summary

Appmeta provides users to upload application metadata and ability to search the uploaded 
metadata.

## Build Requirements
* golang (>= 1.12.0)
* make
* docker (optional)

## How to build

### Using Make

* Running unit test: __make test__
* Building application: __make build__. which will produce the executable in bin/appmeta
* Running application: __./bin/appmeta -addr=localhost:8080 -conf=./conf__

### Docker

* docker build -t appmeta/latest .
* docker run -p 8080:8080 appmeta/latest


##Configuration

The conf/analyzer.json defines the fieldName to tokenizer mapping  which can be overriden. There are four types of tokenizers that are currently 
supported:
1. StandardTokenizer: 
	1. Converts the input to lowercase
	2. Splits the input into tokens based on the separator
	3. Trims the tokens based on the cutset specified
	4. Filters the stopWords out of the tokens
2. ExactMatchTokenizer:
	1. Does not do anything except for converting the input to lowercase
3. TokenizerChain:
	1. Makes a Tokenizer by combining 2 or more tokenizers
4. NopTokenizer:
	1. Does not emit anything useful if the field not be indexed.
	


## API Endpoints Summary

Description |Endpoint | Request | Response    |
------------|---------|-------------|-------------|
Index metadata | POST /api/v1/metadata | Metadata Object in body | 201 on success with uuid in the Location header, 400 on validation errors|
Search metadata| GET  /api/v1/metadata/_search | search filters as query params | List of Metadata objects that matched the query |
Get all metadata| GET  /api/v1/metadata  | None | List of all Metadata objects |
Get metadata   | GET  /api/v1/metadata/{uuid}  | UUID as path param | Metadata object with the given ID |
Get service health | GET /api/v1/metadata/health | None | Health status |
Get stats | GET /api/v1/stats | None | service Stats (expvar)

## Important Endpoints Details

1. POST /api/v1/metadata

Indexes and stores the given metadata and returns 201 if there are no validation errors. The supported mime types are 'application/x-yaml' and the
client has to set the 'Content-Type' header to application/x-yaml otherwise server might reject the response. During the
indexing call the some of the fields are converted to lowercase and broken down to tokens to enable searching while some 
fields like license, website etc are indexed after lowercasing them without any tokenization. Some special fields like 
name are indexed both as tokenized and as exact string to enable searching exact name or firstname or lastname. All the
tokenized fields are filtered for common stop words like "and, an, then this the" etc before indexing.

The following table describes analysis done for each field before indexing

Field | Lowercased | Exact Match | Tokenized(And stop words filtered) | Description |
------|------------|-------------|------------------------------------|-------------|
Title | Yes | Yes | No | Search on this field can be an exact search only |
Name | Yes | Yes | Yes | Search can be on firstname, lastname or full name |
Email| Yes | Yes | No | Search has to be exact match |
Version| Yes | Yes | No | Search has to be exact match |
Company | Yes | Yes | No | Search has to be exact match |
Website | Yes | Yes | No | Search has to be exact match |
Source | Yes | Yes | No | Search has to be exact match |
License | Yes | Yes | No | Search has to be exact match |
Description | Yes | No | Yes | Search has to be on individual words that are not stopwords like "the, and" etc |

e.g. index call 
```shell
curl -v -XPOST -H "Content-Type: application/x-yaml" localhost:8080/api/v1/metadata -d 'title: Valid App 2
version: 1.0.1
maintainers:
- name: AppTwo Maintainer
  email: apptwo@hotmail.com
company: Feye Inc.
website: https://upbound.io
source: https://github.com/upbound/repo
license: Apache-2.0
description: |
 ### Why app 2 is the best
 Because it simply is...
'
Note: Unnecessary use of -X or --request, POST is already inferred.
*   Trying ::1...
* TCP_NODELAY set
* Connected to localhost (::1) port 8080 (#0)
> POST /api/v1/metadata HTTP/1.1
> Host: localhost:8080
> User-Agent: curl/7.54.0
> Accept: */*
> Content-Type: application/x-yaml
> Content-Length: 278
> 
* upload completely sent off: 278 out of 278 bytes
< HTTP/1.1 201 Created
< Location: /api/v1/metadata/ca17446c-4aa6-11e9-8e13-f40f2410afb9
< Date: Wed, 20 Mar 2019 00:26:22 GMT
< Content-Length: 0
< 
* Connection #0 to host localhost left intact

``` 

2. GET /api/v1/metadata/_search?name=term&company=term2

Search endpoint returns the list of metadata objects that match the given search filters. The search filters are specified as the
queryparams where only one value per searchfield is supported. The Searchfields directly map to the different fields of the 
metadata object so the search fields that are supported are name, title, company, description etc. There is one __extra__
search field called __any__ that can be used to specify any field match. If multiple filters are specified they behave as an
&& operation - the search hits match all the filters specified. When no query parameters/filters are supplied the call behaves like 
a getall.

e.g. 1 filter search which results in 2 hits as both payloads match 'because' in the description field
```shell
curl "127.0.0.1:8080/api/v1/metadata/_search?any=because"
- _id: 7ac74f86-4ab2-11e9-a15f-f40f2410afb9
  metadata:
    title: Valid App 2
    version: 1.0.1
    maintainers:
    - name: AppTwo Maintainer
      email: apptwo@hotmail.com
    company: Feye Inc.
    website: https://feye.io
    source: https://github.com/feye/repo
    license: Apache-2.0
    description: |
      ### Why app 2 is the best
      Because it simply is...
- _id: 86446f60-4ab2-11e9-a15f-f40f2410afb9
  metadata:
    title: Valid App 2
    version: 1.0.1
    maintainers:
    - name: V Poliboyina
      email: apptwo@hotmail.com
    company: Feye Inc.
    website: https://feye.io
    source: https://github.com/feye/repo2
    license: Apache-2.0
    description: |
      ### Why app 2 is the best
      Because it is awesome

```

e.g. 2 filter search which gives a 1 hit as both the descriptions match 'because'
```shell
curl -H "Content-Type: application/x-yaml" "127.0.0.1:8080/api/v1/metadata/_search?any=because&name=poliboyina"
- _id: 86446f60-4ab2-11e9-a15f-f40f2410afb9
  metadata:
    title: Valid App 2
    version: 1.0.1
    maintainers:
    - name: V Poliboyina
      email: apptwo@hotmail.com
    company: Feye Inc.
    website: https://feye.io
    source: https://github.com/feye/repo2
    license: Apache-2.0
    description: |
      ### Why app 2 is the best
      Because it is awesome


```



