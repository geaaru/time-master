module github.com/geaaru/time-master

go 1.16

replace github.com/docker/docker => github.com/Luet-lab/moby v17.12.0-ce-rc1.0.20200605210607-749178b8f80d+incompatible

replace github.com/opencontainers/runc => github.com/opencontainers/runc v1.0.0-rc9.0.20200221051241-688cf6d43cc4

require (
	github.com/MottainaiCI/mottainai-server v0.1.1
	github.com/kyokomi/emoji v2.2.4+incompatible
	github.com/logrusorgru/aurora v2.0.3+incompatible
	github.com/olekukonko/tablewriter v0.0.5
	github.com/onsi/ginkgo/v2 v2.1.1
	github.com/onsi/gomega v1.17.0
	github.com/rickb777/date v1.16.1
	github.com/spf13/afero v1.8.0 // indirect
	github.com/spf13/cobra v1.2.1
	github.com/spf13/viper v1.10.1
	go.uber.org/zap v1.19.1
	golang.org/x/sys v0.0.0-20220128215802-99c3d69c2c27 // indirect
	gopkg.in/ini.v1 v1.66.3 // indirect
	gopkg.in/yaml.v2 v2.4.0
)
