module github.com/superchalupa/sailfish

go 1.27.0

require (
	github.com/Knetic/govaluate v3.0.1-0.20171022003610-9aa49832a739+incompatible
	github.com/dgrijalva/jwt-go v3.2.0+incompatible
	github.com/etcd-io/bbolt v1.3.2
	github.com/fsnotify/fsnotify v1.4.7
	github.com/gorilla/mux v1.6.2
	github.com/inconshreveable/log15 v0.0.0-20180818164646-67afb5ed74ec
	github.com/json-iterator/go v1.1.11
	github.com/looplab/eventhorizon v0.4.0
	github.com/mitchellh/mapstructure v1.0.0
	github.com/spacemonkeygo/openssl v0.0.0-20180913232302-a66df3e4f582
	github.com/spf13/pflag v1.0.10
	github.com/spf13/viper v1.2.0
	github.com/stretchr/testify v1.11.1
)

require (
	github.com/BurntSushi/toml v0.3.1 // indirect
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/go-stack/stack v1.8.0 // indirect
	github.com/gorilla/context v1.1.1 // indirect
	github.com/hashicorp/hcl v1.0.0 // indirect
	github.com/kr/pretty v0.1.0 // indirect
	github.com/magiconair/properties v1.8.0 // indirect
	github.com/mattn/go-colorable v0.0.9 // indirect
	github.com/mattn/go-isatty v0.0.4 // indirect
	github.com/modern-go/concurrent v0.0.0-20180306012644-bacd9c7ef1dd // indirect
	github.com/modern-go/reflect2 v1.0.1 // indirect
	github.com/pelletier/go-toml v1.2.0 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/spacemonkeygo/spacelog v0.0.0-20180420211403-2296661a0572 // indirect
	github.com/spf13/afero v1.1.2 // indirect
	github.com/spf13/cast v1.2.0 // indirect
	github.com/spf13/jwalterweatherman v1.0.0 // indirect
	go.etcd.io/bbolt v1.5.0 // indirect
	golang.org/x/sys v0.45.0 // indirect
	golang.org/x/text v0.3.2 // indirect
	gopkg.in/check.v1 v1.0.0-20190902080502-41f04d3bba15 // indirect
	gopkg.in/yaml.v2 v2.3.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

// replace github.com/looplab/eventhorizon => github.com/superchalupa/eventhorizon v0.0.1

replace github.com/looplab/eventhorizon => github.com/looplab/eventhorizon v0.2.1-0.20180328082012-7067a22f516d
