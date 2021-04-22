#!/bin/bash

rm -fr download

jq -ncM 'while(true; .+1) | {method: "POST", url: "http://localhost:8080/download", body: {key1:"value1"} | @base64 }' | \
vegeta attack -rate=2/s -lazy -format=json -duration=5s > results.2qps.bin 