import { DataSourceInstanceSettings } from '@grafana/data';
import { DataSourceWithBackend, getBackendSrv } from '@grafana/runtime';
import { WebMonitoringDataSourceOptions, WMResultsQuery, WebMonitoringMonitor } from './types';

export class DataSource extends DataSourceWithBackend<WMResultsQuery, WebMonitoringDataSourceOptions> {
  constructor(instanceSettings: DataSourceInstanceSettings<WebMonitoringDataSourceOptions>) {
    super(instanceSettings);
  }

  // Enables default annotation support for 7.2+
  annotations = {};

  async getMonitors(): Promise<WebMonitoringMonitor[]> {
    let rawMonitors = await getBackendSrv().get(`/api/datasources/${this.id}/resources/rm/webmonitoring/monitors`);

    let monitors: WebMonitoringMonitor[] = [];

    if (!rawMonitors) {
      return monitors;
    }

    rawMonitors.forEach((element: any) => {
      let monitor: WebMonitoringMonitor = {
        id: element.monitorId,
        type: element.type,
        name: element.name,
        url: element.url,
      };

      monitors.push(monitor);
    });
    return monitors;
  }
}
