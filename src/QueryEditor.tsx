import React, { PureComponent } from 'react';
import { InlineField, Select, LegacyForms } from '@grafana/ui';
import { QueryEditorProps, SelectableValue } from '@grafana/data';
import { DataSource } from './DataSource';
import {
  WebMonitoringDataSourceOptions,
  WMResultsQuery,
  ProductType,
  QueryTypeValue,
  WebMonitoringMonitor,
} from './types';
const { FormField } = LegacyForms;

const defaultQueryProduct: ProductType = 'webmonitoring';
const defaultQueryType: QueryTypeValue = 'monitorresults';

const queryTypeOptions: Array<SelectableValue<QueryTypeValue>> = [
  { value: 'monitorresults', label: 'Monitor Results' },
  { value: 'monitors', label: 'Monitors (Table)' },
  { value: 'alarms', label: 'Alarms (Table)' },
];

type Props = QueryEditorProps<DataSource, WMResultsQuery, WebMonitoringDataSourceOptions>;

interface Istate {
  monitors: Array<SelectableValue<string>>;
}

export class QueryEditor extends PureComponent<Props, Istate> {
  constructor(props: Props) {
    super(props);

    this.state = {
      monitors: [],
    };
  }

  componentDidMount() {
    this.fetchMonitorsFromDataSource();
  }

  componentWillMount() {
    if (!this.props.query.hasOwnProperty('queryProduct')) {
      this.props.query.queryProduct = defaultQueryProduct;
    }
    if (!this.props.query.hasOwnProperty('queryType')) {
      this.props.query.queryType = defaultQueryType;
    }
  }

  async fetchMonitorsFromDataSource() {
    const { datasource } = this.props;

    const requestedMonitors = await datasource.getMonitors();
    const selectableMonitors = requestedMonitors.map((monitor: WebMonitoringMonitor) => {
      return this.makeWebMonitoringMonitorSelectable(monitor);
    });

    this.setState({
      monitors: selectableMonitors,
    });
  }

  onMonitorChange = (selectedMonitor: SelectableValue<string>) => {
    const { query, onRunQuery, onChange } = this.props;
    if (!selectedMonitor) {
      return; // ignore delete?
    }

    onChange({
      ...query,
      queryMonitorId: selectedMonitor.id,
      queryMonitorDetails: {
        name: selectedMonitor.name,
        type: selectedMonitor.type,
        url: selectedMonitor.url,
      },
    });
    onRunQuery();
  };

  onQueryTypeChange = (selectedQueryType: SelectableValue<QueryTypeValue>) => {
    const { query, onRunQuery, onChange } = this.props;

    if (selectedQueryType.value) {
      onChange({
        ...query,
        queryType: selectedQueryType.value,
      });
      onRunQuery();
    }
  };

  makeWebMonitoringMonitorSelectable = (monitor: WebMonitoringMonitor): SelectableValue<string> => {
    return {
      ...monitor,
      value: monitor.name ? monitor.name : '',
      label: monitor.name ? monitor.name : '',
    };
  };

  renderMonitorResultsInputForm = () => {
    if (this.props.query.queryType !== 'monitorresults') {
      return;
    }

    const monitorValue = this.props.query.queryMonitorDetails
      ? this.makeWebMonitoringMonitorSelectable({
          ...this.props.query.queryMonitorDetails,
          id: this.props.query.queryMonitorId,
        })
      : '';

    return (
      <>
        <div className="gf-form-inline max-width-30">
          <InlineField label="Monitors" tooltip="Available monitors" grow={true} labelWidth={14}>
            <Select
              options={this.state.monitors}
              value={monitorValue}
              onChange={this.onMonitorChange}
              menuPlacement={'bottom'}
              placeholder="Select one Monitor"
              width={24}
            />
          </InlineField>
        </div>
        <div className="gf-form max-width-30">
          <FormField
            labelWidth={7}
            value={this.props.query.queryMonitorDetails ? this.props.query.queryMonitorDetails.type : ''}
            label="Monitor Type"
            tooltip="Monitor Type configured {Icmp, Http, Https, PageLoad, Transaction}"
            contentEditable={false}
            disabled={true}
            width={25}
          />
        </div>
        <div className="gf-form max-width-30">
          <FormField
            labelWidth={7}
            value={this.props.query.queryMonitorDetails ? this.props.query.queryMonitorDetails.url : ''}
            label="Monitor URL"
            tooltip="Monitor URL configured"
            contentEditable={false}
            disabled={true}
            width={25}
          />
        </div>
      </>
    );
  };

  render() {
    return (
      <>
        <div className="gf-form-inline max-width-30">
          <InlineField label="Query Type" tooltip="Available query types" grow={true} labelWidth={14}>
            <Select
              options={queryTypeOptions}
              value={this.props.query.queryType}
              onChange={this.onQueryTypeChange}
              menuPlacement={'bottom'}
              placeholder="Select one Query Type"
              width={24}
            />
          </InlineField>
        </div>
        {this.renderMonitorResultsInputForm()}
      </>
    );
  }
}
