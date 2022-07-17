# opentelemetry

Feature
---
- Tracing 
  - support tracing gorm by Hook `Create` `Query` `Delete` `Update` `Row` `Raw` 
- Metrics 
  - Collect DB Status 
- Logrus
  - Use logrus replace gorm default logger
  - Use hook to report span message
- Provider
  - Out-of-the-box default opentelemetry provider
  - Support setting via environment variables
---
Set logger
~~~go
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
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{Logger: logger})
}
~~~
Set tracing and metrics
~~~go
import(
	"gorm.io/plugin/opentelemetry/tracing"
)
func init(){

	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{Logger: logger})
	if err != nil {
		panic(err)
	}

	if err := db.Use(tracing.NewPlugin()); err != nil {
		panic(err)
	}
}
~~~