import { DataSourcePlugin } from '@grafana/data';
import { DataSource } from './DataSource';
import { ConfigEditor } from './ConfigEditor';
import { QueryEditor } from './QueryEditor';
import { WMResultsQuery, WebMonitoringDataSourceOptions } from './types';

export const plugin = new DataSourcePlugin<DataSource, WMResultsQuery, WebMonitoringDataSourceOptions>(DataSource)
  .setConfigEditor(ConfigEditor)
  .setQueryEditor(QueryEditor);
