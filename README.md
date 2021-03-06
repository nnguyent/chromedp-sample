# Test chromedp can download multiple simultaneous request

## Quick start

This sample used [vegeta](https://github.com/tsenart/vegeta) to test chromedp can download multiple simultaneous request.

1. To install vegeta, run the command:

    ```bash
    go get -u github.com/tsenart/vegeta
    ```

2. This sample will start http service and listen on `localhost:8080`

## Run the sample directly from source code

1. Run main.go in terminal 1:

    ```bash
    go run main.go
    ```

2. Run script attack.sh in terminal 2:

    ```bash
    ./attack.sh
    ```

## Run the sample in docker container

1. Build/start docker container in terminal 1:

    ```bash
    docker-compose up -d --build
    docker logs -f chromedp_sample 
    ```

    - To stop the container `chromedp_sample`, run command:

    ```bash
    docker-compose down 
    ```

2. Run script attack.sh in terminal 2:

    ```bash
    ./attack.sh
    ```
