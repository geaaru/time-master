# Time Master

[![Build Status](https://travis-ci.com/geaaru/time-master.svg?branch=master)](https://travis-ci.com/geaaru/time-master)
[![Go Report Card](https://goreportcard.com/badge/github.com/geaaru/time-master)](https://goreportcard.com/report/github.com/geaaru/time-master)

All you need to monitor your activities and to plan your team tasks with an Opensource project.

Based on the idea of [TaskJuggler](https://taskjuggler.org/) project.


## Getting Start

### Import timesheet from JIRA

Hereinafter, an example about how you can import a Jira Timesheet Report and convert to `time-master` yaml specs.

```
Import CSV timesheet

Usage:
   import timesheet [file] [flags]

Flags:
  -d, --dir string                Directory where import timesheets.
  -h, --help                      help for timesheet
  -i, --import-type string        Define type of the imported file. Now it's supported only Jira. (default "jira")
  -j, --jira-mapper-file string   Import jira resource mapper file.
  -s, --split-for-user            Create a timesheet file for every user.
      --stdout                    Print timesheets to stdout instead of write files.
  -p, --target-prefix string      Prefix of the file/files to create.

Global Flags:
  -c, --config string   Time Master configuration file
  -v, --verbose         Verbose output.
```

The *jira-mapper-file* is used for map JIRA user to Time Master users and to assign Jira issues to a specific task.

An example of jira mapper file:

```yaml
resources:
- jira_name: "Daniele Rondina"
  name: "geaaru"

issues:
- jira_issue: "ISSUE-1"
  task: "MYCLIENT01.briefing"

```

Split import with a file per user:

```shell

$> time-master import timesheet Reports_2020-06-01_2020-06-30.csv -d workspace/timesheets/202006/ -j workspace/mapper/jira.yaml  -s

```

Import to a single file.

```shell

$> time-master import timesheet Reports_2020-06-01_2020-06-30.csv -d workspace/timesheets/202006/ -j workspace/mapper/jira.yaml

```


