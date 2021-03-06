[![Build Status](https://github.com/LuckyLukas/whenthengo/workflows/check-build/badge.svg)](https://github.com/LuckyLukas/whenthengo/actions) [![Release Integrationtest Status](https://github.com/LuckyLukas/whenthengo/workflows/release/badge.svg)](https://github.com/LuckyLukas/whenthengo/actions)
![codecov](https://codecov.io/gh/LuckyLukas/whenthengo/branch/master/graph/badge.svg)
# whenthengo

a simple mock http server to use for testing with packages like
[testcontainers-go](https://github.com/testcontainers/testcontainers-go).`

<b>looking for the client?</b> [here you go](https://github.com/LuckyLukas/whenthengo/tree/master/client)


# Table of Contents
1. [what?](#what)
2. [limitations](#limitations)
3. [whens](#whens)
4. [thens](#thens)
5. [api](#api)
6. [matching](#matching)
7. [config format](#formats)

## what? <a name="what"></a>

whenthengo is configured using _whens_ and _thens_, if the app recognizes an http request to match a _when's_ conditions, it will response with the contents present in the matching _then_.

## limitations for 1.0.0 <a name="limitations"></a>

- no support for fuzzy matches or best effort responses, 
- _whens_ that depend on whitespace positions or case in headers or body may fail to match

### whens <a name="whens"></a>

a when is a set of conditions a request must meet to make the server return the matching _then_.

### thens <a name="thens"></a>

are linked to one _when_ and describe the expected response.

## api parameters <a name="api"></a>

### whens

| property        | type           | desciption  |
| ------------- |-------------| -----|
| method     | string| case insensitive http verb to match|
| url     | string      |   the path to match |
| headers | map[string]anything       |    a map of headers the request should include. This checks for "containment" and is case insensitive (eg.: ```"application/json; encoding=UTF-8"``` with match ```"Application/json"```). We tried to be clever with parsing strings, arrays and numerics, as long as the headers are somehow resembling a real situation, it should work |
| body| string | the request body to match. this will remove all whitespaces when checking for equiality.|

### thens
| property        | type           | desciption  |
| ------------- |-------------| -----|
| delay     | int| artificial delay for the response in milliseconds |
| status     | int      |   http status code |
| headers | map[string]anything      |    a map of headers the response will include, we tried to be clever, as long as the headers are somehow resembling a real situation, it should work |
| body| string | the expected response body|


### api /whenthengo

#### GET /up
returns an empty ```200 OK```.
 
This can be used to check if the app is up and ready to handle requests.

#### POST /whenthen

post a whenthen as JSON in the request body, just as you would with the configuration file.
Whenthengo will add it to existing configured whenthens.

We also provide a client for that under [here](github.com/luckylukas/whenthengo/client)

### matching headers <a name="matching"></a>

Matching headers is case insensitive.
any spaces in header keys are dropped, so are any spaces, "," or ";" in header values for ease of matching.

Whenthengo will ignore superfluous headers in the request (which don't match any key in the when setup),
but insistns on having matches for all headers present in the _when_.

#### examples

| when.headers        | request.headers           | match  |
| ----------- |------------------ | ------------- |
| key= ["A"]       | key= ["a", "b"]              | true |
| key= []       | key= ["a", "b"]                 | true |
| key= ["a"]       | key= ["a;"]              | true |
| key= ["a", "b"]       | key= ["a"]              | false |
| key= ["a"]       | key= ["a;encoding=UTF-8"]              | false |


### matching body

the body is simply matched by removing any whitespace (```\r,\n, ,\t```) from the configured and the requested bodies and checking for equality.

| when.body        | request.body           | match  |
| ----------- |------------------ | ------------- |
| { "data":"a"}      | {\t"data":"a"\n}            | true |
| {"data": "hello world"}      | {"data":"helloworld"}            | true |


### conflicts

if two _whens_ collide, the first one to match will be returned.

for example:
```json
[
{"when": {"url":  "/", "method": "post", "headers": {"accept": ["text", "octet-stream"]}}, "then":  {}}
{"when": {"url":  "/", "method": "post", "headers": {"accept": ["text"]}}, "then":  {}}
]
```

requesting ```POST / accept="text"``` will match the first one return it's then.


## input formats <a name="formats"></a>

currently you can specify when-thens in ```yaml```
 and ```json``` format

### json

```json
[
  {
    "when": {
      "method": "get",
      "url": "/path/test",
      "headers": {
        "Accept": "application/json",
        "Content-Length": "6"
      },
      "body": "some body"
    },
    "then": {
      "delay": 100,
      "status": 200,
      "headers": {
        "Accept": ["1", "2"]
      },
      "body": "k"
    }
  },
  {}
]
```

### yaml

```
    -
      when:
        method: "get"
        url: "/path/test"
        headers:
          Accept: "application/json"
          Content-Length: 6
        body: |+
          abc
          def
      then:
        delay: 100
        status: 200
        headers:
          Accept: 
          - 1
          - 2
        body: "k"
    -
        ...

```

## bugs and collaboration

yes and please.
