watchdog-lambda
=========

[![GitHub Workflow Status](https://img.shields.io/github/actions/workflow/status/akerl/watchdog-lambda/build.yml?branch=main)](https://github.com/akerl/watchdog-lambda/actions)
[![GitHub release](https://img.shields.io/github/release/akerl/watchdog-lambda.svg)](https://github.com/akerl/watchdog-lambda/releases)
[![MIT Licensed](https://img.shields.io/badge/license-MIT-green.svg)](https://tldrlegal.com/license/mit-license)

Lambda for alerting if actions don't occur in a timely manner

## Usage

The Lambda expects a config file that looks like this:

```
slack_webhook: https://your-slack-webhook
checks:
  - name: backup-s3
    key: 7AC5458F-7473-4470-9207-4CCF40D0CD9E
    threshold: 1500
  - name: backup-github
    key: CF486E68-E4FD-4969-B4A8-D9C40019A978
    threshold: 3600
```

Then, to check in for `backup-s3`, just `curl https://your-lambda-domain/checks/7AC5458F-7473-4470-9207-4CCF40D0CD9E`

The keys don't need to be UUIDs, it's just a relatively easy way to generate semi-random keys. Threshold is measured in minutes.

## Installation

The easiest way to deploy this is probably to use [the Terraform module](https://registry.terraform.io/modules/armorfret/lambda-watchdog/aws/latest?tab=inputs). The module creates a bucket to hold the configuration file; you'll need to write the config to a file called `config.yaml` in the root of the bucket.

## License

watchdog-lambda is released under the MIT License. See the bundled LICENSE file for details.
