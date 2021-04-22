#!/bin/bash

rm -fr download

jq -ncM 'while(true; .+1) | {method: "POST", url: "http://localhost:8080/download", body: {key1:"value1"} | @base64 }' | \
vegeta attack -rate=3/s -lazy -format=json -duration=3s > results.2qps.bin 