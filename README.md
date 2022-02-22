# OnlyOffice Exporter for Prometheus

Based on the work from https://edu-git.ac-versailles.fr/prometheus/exporter/onlyoffice.git

Exports OnlyOffice statistics via HTTP for Prometheus consumption.

Help on flags:

<pre>
Usage of ./prometheus_onlyoffice_exporter:
  -insecure
        Ignore onlyoffice server certificate if using https.
  -scrape_uri string
        URI to the onlyoffice statistics info. (default "http://localhost/info/info.json")
  -web.listen-address string
        Address on which to expose metrics. (default ":9999")
  -web.telemetry-path string
        Path under which to expose metrics. (default "/metrics")
</pre>

Pay attention that by default, the `info/info.json` page is only available from localhost.
The onlyoffice's nginx configuration has to be modified if the exporter is not running localy.

This modified version is including the number of active connections, valid to monitor the enterprise licenses left. `onlyoffice_connections_current`.

## Collectors

OnlyOffice metrics:

```
# HELP onlyoffice_view_connections_last_hour Number of view connections during last hour
# TYPE onlyoffice_view_connections_last_hour gauge
# HELP onlyoffice_edit_connections_last_hour Number of edit connections during last hour
# TYPE onlyoffice_edit_connections_last_hour gauge
# HELP onlyoffice_view_connections_last_day Number of view connections during last day
# TYPE onlyoffice_view_connections_last_day gauge
# HELP onlyoffice_edit_connections_last_day Number of edit connections during last day
# TYPE onlyoffice_edit_connections_last_day gauge
# HELP onlyoffice_view_connections_last_week Number of view connections during last week
# TYPE onlyoffice_view_connections_last_week gauge
# HELP onlyoffice_edit_connections_last_week Number of edit connections during last week
# TYPE onlyoffice_edit_connections_last_week gauge
# HELP onlyoffice_view_connections_last_month Number of view connections during last month
# TYPE onlyoffice_view_connections_last_month gauge
# HELP onlyoffice_edit_connections_last_month Number of edit connections during last month
# TYPE onlyoffice_edit_connections_last_month gauge
# HELP onlyoffice_license_info License Information on OnlyOffice
# TYPE onlyoffice_license_info gauge
# HELP onlyoffice_server_info Server Information of OnlyOffice
# TYPE onlyoffice_server_info gauge
# HELP onlyoffice_up Could the OnlyOffice server be reached
# TYPE onlyoffice_up gauge
# HELP onlyoffice_connections_current Number of connections currently active
# TYPE onlyoffice_connections_current gauge

```

