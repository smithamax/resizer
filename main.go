package main

import (
	"fmt"
	"net/http"

	"github.com/smithamax/resizer/resizer"
	"github.com/spf13/viper"
)

func main() {
	viper.SetConfigName("config")
	viper.AddConfigPath("/etc/resizer/")
	viper.AddConfigPath(".")
	viper.SetEnvPrefix("resizer")
	viper.AutomaticEnv()

	err := viper.ReadInConfig()
	if err != nil {
		panic(fmt.Errorf("Fatal error config file: %s \n", err))
	}

	var source resizer.Source

	switch viper.GetString("source_type") {
	case "filesystem":
		path := viper.GetString("filesystem_source_path")
		source = resizer.FileSource{Root: path}
	case "s3":
		bucket := viper.GetString("s3_source_bucket")
		region := viper.GetString("s3_source_region")
		prefix := viper.GetString("s3_source_prefix")
		source, err = resizer.NewS3Source(bucket, region, prefix)
	}
	if source == nil {
		panic(fmt.Errorf("source_type not configured"))
	}

	if err != nil {
		panic(err)
	}

	resizerHandler := resizer.Handler(source)
	http.Handle("/images/", http.StripPrefix("/images/", resizerHandler))
	http.Handle("/images", http.StripPrefix("/images", resizerHandler))
	http.ListenAndServe(":8080", nil)
}
