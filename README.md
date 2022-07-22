# opentelemetry
[Opentelemetry](https://opentelemetry.io/) for [gorm](https://github.com/go-gorm/gorm)
## Feature 
### Tracing 
  - support tracing gorm by Hook `Create` `Query` `Delete` `Update` `Row` `Raw` 
### Metrics 
  - Collect DB Status
### Logging
  - Use logrus replace gorm default logger
  - Use hook to report span message
### Provider
  - Out-of-the-box default opentelemetry provider
  - Support setting via environment variables

## How to Use ?
### Set logger
~~~go
package main

import(
	"gorm.io/gorm/logger"
	"gorm.io/plugin/opentelemetry/logging/logrus"
)

func init(){
	logger := logger.New(
		logrus.NewWriter(),
		logger.Config{
			SlowThreshold: time.Millisecond,
			LogLevel:      logger.Warn,
			Colorful:      false,
		},
	)
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"),&gorm.Config{Logger: logger})
}
~~~
### Set tracing and metrics

~~~go
package main

import(
	"gorm.io/plugin/opentelemetry/tracing"
)

func init(){

	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	if err != nil {
		panic(err)
	}

	if err := db.Use(tracing.NewPlugin()); err != nil {
		panic(err)
	}
}
~~~

### Set only tracing
~~~go
package main

import(
	"gorm.io/plugin/opentelemetry/tracing"
)

func init(){

	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	if err != nil {
		panic(err)
	}

	if err := db.Use(tracing.NewPlugin(tracing.WithoutMetrics())); err != nil {
		panic(err)
	}
}
~~~



### More info
See [examples](examples/)