## RedisPOC
* It is a basic implementation of redis-go library functions for storing and retrieving redis cache.
* Using redis cache reduces the execution time of searching and retrieving entries when compared to any other database.
* Observation
* ![observation](/observation.png)
* As shown in above observation the first successfull request took around 105ms whereas second request which is fetched from redis cache took only around 17ms.