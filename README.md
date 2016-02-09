
This is a backend component for [ToDoLite-iOS](https://github.com/couchbaselabs/ToDoLite-iOS) which has the following commands:

* Webserver which dumps the [changes feed](http://couchbase-mobile.s3.amazonaws.com/misc/issue_1526_test_fest_no_sync/changes.html) 
* Follows the changes feed of the TodoLite sync gateway database and whenever a new image is uploaded, it runs it through OCR and saves the decoded text into the JSON.

## How to use this

```
$ export GO15VENDOREXPERIMENT=1
$ go get -u github.com/tleyden/todolite-appserver
$ todolite-appserver --help
```



