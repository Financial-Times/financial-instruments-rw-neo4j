# Financial Instruments Reader/Writer for Neo4j (financial-instruments-rw-neo4j)

[![Circle CI](https://circleci.com/gh/Financial-Times/financial-instruments-rw-neo4j/tree/master.png?style=shield)](https://circleci.com/gh/Financial-Times/financial-instruments-rw-neo4j/tree/master)
[![Go Report Card](https://goreportcard.com/badge/github.com/Financial-Times/financial-instruments-rw-neo4j)](https://goreportcard.com/report/github.com/Financial-Times/financial-instruments-rw-neo4j)


__An API for reading and writing financial instruments into Neo4j. This service expects that the financial instruments received to be in the format used by financial instruments transformer.__

#Installation

For the first time:

`go get github.com/Financial-Times/financial-instruments-rw-neo4j`

or update:

`go get -u github.com/Financial-Times/financial-instruments-rw-neo4j`

#Running

`$GOPATH/bin/financial-instruments-rw-neo4j --neo-url={neo4jUrl} --port={port} --batchSize=50 --graphiteTCPAddress=graphite.ft.com:2003 --graphitePrefix=content.{env}.financial-instruments.rw.neo4j.{hostname} --logMetrics=false

All arguments are optional, they default to a local Neo4j install on the default port (7474), application running on port 8080, batchSize of 1024, graphiteTCPAddress of "" (meaning metrics won't be written to Graphite), graphitePrefix of "" and logMetrics false.

NB: the default batchSize is much higher than the throughput the instance data ingester currently can cope with. 

## Updating the model

We use the transformer to get the information to write and from that we establish the json for the request. This representation is held in the model.go in a struct called financialInstrument.

Use gojson against a transformer endpoint to create a financialInstrument struct and update the financialInstrument/model.go file. 

<!--todo add financial instruments transformer url-->
`curl http://ftaps39403-law1a-eu-t:8080/`

Response format for current version:

`{
    "uuid": "6562674e-dbfa-4cb0-85b2-41b0948b7cc2",
    "prefLabel": "Contract",
    "alternativeIdentifiers":{
        "uuids":[
            "6562674e-dbfa-4cb0-85b2-41b0948b7cc2"
        ],
        "factsetIdentifier": "B000BB-S",
        "figiCode": "BBG000Y1HJT8"
    },
    "issuedBy": "4e484678-cf47-4168-b844-6adb47f8eb58"
 }`

## Endpoints

/financialInstruments/{uuid}

### PUT
The only mandatory field is the uuid, and the alternativeIdentifier uuids (because the uuid is also listed in the alternativeIdentifier uuids list). The uuid in the body must match the one used on the path.

Every request results in an attempt to update that financial instrument using the Neo4j MERGE clause, which updates the pattern if it exists, otherwise creates a new one.

A successful PUT results in 200.

We run queries in batches. If a batch fails, all failing requests will get a 500 server error response.

Example PUT request:

    `curl -XPUT localhost:8080/financialInstruments/6562674e-dbfa-4cb0-85b2-41b0948b7cc2 \
         -H "X-Request-Id: 123" \
         -H "Content-Type: application/json" \
         -d `{"uuid":"6562674e-dbfa-4cb0-85b2-41b0948b7cc2","prefLabel":"GREENWICH CAP ACCEPTANCE  1991-B B1","alternativeIdentifiers":{"uuids":["6562674e-dbfa-4cb0-85b2-41b0948b7cc2"],"factsetIdentifier":"B000BB-S","figiCode":"BBG000Y1HJT8"},"issuedBy":"4e484678-cf47-4168-b844-6adb47f8eb58"}`

Invalid json body input, or uuids that don't match between the path and the body will result in a 400 bad request response.

### GET
This internal read should return what got written (i.e., this isn't the public financial instrument read API)

If not found, you'll get a 404 response.

Empty fields are omitted from the response.
`curl -H "X-Request-Id: 123" localhost:8080/financialInstruments/6562674e-dbfa-4cb0-85b2-41b0948b7cc2`

### DELETE
Will return 204 if successful, 404 if not found
`curl -XDELETE -H "X-Request-Id: 123" localhost:8080/financialInstruments/6562674e-dbfa-4cb0-85b2-41b0948b7cc2`

### Admin endpoints
Health checks: http://localhost:8080/__health

Ping: http://localhost:8080/ping or http://localhost:8080/__ping


### Logging
 The application uses logrus, the logfile is initialised in main.go. Logging requires an env app parameter, for all environments  other than local logs are written to file
 When running locally logging is written to console (if you want to log locally to file you need to pass in an env parameter that is != local)
 NOTE: build-info end point is not logged as it is called every second from varnish and this information is not needed in  logs/splunk