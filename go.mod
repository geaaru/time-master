module github.com/geaaru/time-master

go 1.14

replace github.com/docker/docker => github.com/Luet-lab/moby v17.12.0-ce-rc1.0.20200605210607-749178b8f80d+incompatible

require (
	github.com/MottainaiCI/mottainai-server v0.0.0-20200319175456-fc3c442fd4a6
	github.com/fsouza/go-dockerclient v1.6.5 // indirect
	github.com/jaypipes/ghw v0.6.1 // indirect
	github.com/kyokomi/emoji v2.2.4+incompatible
	github.com/logrusorgru/aurora v0.0.0-20190417123914-21d75270181e
	github.com/mudler/luet v0.0.0-20200717204249-ffa6fc3829d2
	github.com/olekukonko/tablewriter v0.0.4
	github.com/onsi/ginkgo v1.14.0
	github.com/onsi/gomega v1.10.1
	github.com/spf13/cobra v1.0.0
	github.com/spf13/viper v1.7.0
	go.uber.org/zap v1.15.0
	gopkg.in/clog.v1 v1.2.0 // indirect
	gopkg.in/src-d/go-git.v4 v4.13.1 // indirect
	gopkg.in/yaml.v2 v2.3.0
)
