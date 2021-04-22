# Test chromedp can download multiple simultaneous request

## Quick start

This sample used [vegeta](https://github.com/tsenart/vegeta) to test chromedp can download multiple simultaneous request.

1. To install vegeta, run the command:

    ```bash
    go get -u github.com/tsenart/vegeta
    ```

2. This sample will start http service and listen on `localhost:8080`

## Run the sample

1. Run main.go in terminal 1:

    ```bash
    go run main.go
    ```

2. Run script attack.sh in terminal 2:

    ```bash
    ./attack.sh
    ```
