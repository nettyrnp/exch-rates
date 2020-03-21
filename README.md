Exchange Rates Service
========================
Backend part of the Exchange Rates Service.
Exchange Rates Service is a service that collects and stores the currency exchange rates several times per minute. It also provides an API for quering this info.

## Installation

#### Install Go
https://golang.org/doc/install

#### Download the App repository
```
go get github.com/nettyrnp/exch-rates
```

#### Install and Start Postgres
[Ubuntu]
https://www.digitalocean.com/community/tutorials/how-to-install-and-use-postgresql-on-ubuntu-18-04

[MacOS]
```
brew install postgresql
brew services start postgresql
```

Create database in postgres
```
createdb exchrates_be
```

(optional) Try to connect to the database:
```
psql exchrates_be
```

Create tables in the database:
```
make migrate
```

## Running the application
#### Running:
```
make run
```
Now visit http://localhost:8080/api/v0/exchrates/admin/version and see the App version in your browser. 


## REST API:
Examples of Postman requests can be found in testdata/nettyrnp-exchrates.postman_collection.json

#### Main routes:
    GET localhost:8080/api/v0/exchrates/admin/version   // to get the exchange rates API version
    GET localhost:8080/api/v0/exchrates/admin/logs      // to get latest part of logs
    
    POST localhost:8080/api/v0/exchrates/start_poll     // to start gathering of currency exchange rates
    POST localhost:8080/api/v0/exchrates/stop_poll      // to stop gathering of currency exchange rates
    
    GET localhost:8080/api/v0/exchrates/status          // to get the last value of the currency exchange rate, together with the average for 1 day, 1 week, 1 month
    POST localhost:8080/api/v0/exchrates/history        // to get an array of elements that are units of one type of aggregation, the average value of the currency at each moment. Filter options -- time window, aggregation interval (average for 1 min, 5 min, 1 hour, 1 day)
    POST localhost:8080/api/v0/exchrates/momental       // to get the currency exchange rate for the desired moment


#### Sample CURL request:
```
curl -X POST   http://localhost:8080/api/v0/exchrates/start_poll   -H 'cache-control: no-cache'
```
