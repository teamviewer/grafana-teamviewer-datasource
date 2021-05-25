import { DataQuery, DataSourceJsonData } from '@grafana/data';

export interface WMResultsQuery extends DataQuery {
  queryMonitorId: string;
  queryMonitorDetails: WebMonitoringMonitorWithoutId;
  queryProduct: ProductType;
  queryType: QueryTypeValue;
}

export type QueryTypeValue = 'monitorresults' | 'monitors' | 'alarms';

export type ProductType = 'webmonitoring';

/**
 * These are options configured for each DataSource instance
 */
export interface WebMonitoringDataSourceOptions extends DataSourceJsonData {
  path?: string;
}

/**
 * Value that is used in the backend, but never sent over HTTP to the frontend
 */
export interface MySecureJsonData {
  apiToken: string;
}

export interface WebMonitoringMonitorWithoutId {
  type: string;
  name: string;
  url: string;
}

export interface WebMonitoringMonitor extends WebMonitoringMonitorWithoutId {
  id: string;
}
