# TeamViewer Datasource for Grafana ![build status](https://github.com/teamviewer/grafana-teamviewer-datasource/actions/workflows/build.yml/badge.svg)

![build status](https://github.com/teamviewer/grafana-teamviewer-datasource/actions/workflows/golangci-lint.yml/badge.svg)
![build status](https://github.com/teamviewer/grafana-teamviewer-datasource/actions/workflows/codespell.yml/badge.svg)
[![Changelog](https://img.shields.io/badge/change-log-blue.svg?style=flat)](https://github.com/teamviewer/grafana-teamviewer-datasource/blob/master/CHANGELOG.md)
[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](https://raw.githubusercontent.com/teamviewer/grafana-teamviewer-datasource/main/LICENSE)

Visualize your [Teamviewer Web Monitoring](https://www.teamviewer.com/en/remote-management/web-monitoring/) metrics with the leading open source software for time series analytics.

## Installation

Install by using *grafana-cli*,

```bash
grafana-cli plugins install teamviewer
```

Install manually

Download the plugin at your plugin path (i.e. `/var/lib/grafana/plugins`),
then unzip the plugin there with,

```bash
unzip teamviewer-datasource.zip
```

After all those steps restart the Grafana Server service, and you should see the plugin there.

## Configuration

Configure the Datasource with the API Bearer Token (without the word "Bearer"), and save the Datasource.

![](src/img/datasource.png)

Now you can configure a panel on your dashboard as follows,

![](src/img/query.png)

For more information, please refer to the [Wiki](https://github.com/teamviewer/grafana-teamviewer-datasource/wiki) page.

## Contributing

Refer to [CONTRIBUTING.md](https://github.com/teamviewer/grafana-teamviewer-datasource/blob/main/CONTRIBUTING.md)

## License

Apache License 2.0, see [LICENSE](https://github.com/teamviewer/grafana-teamviewer-datasource/blob/main/LICENSE).
