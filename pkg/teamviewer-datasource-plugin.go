package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"sort"
	"strings"
	"time"

	"github.com/grafana/grafana-plugin-sdk-go/backend"
	"github.com/grafana/grafana-plugin-sdk-go/backend/instancemgmt"
	"github.com/grafana/grafana-plugin-sdk-go/backend/log"
	"github.com/grafana/grafana-plugin-sdk-go/data"
)

// webMonitingAPIBasePath base path, see https://webapi.teamviewer.com/api/v1/docs/index#/
const webMonitingAPIBasePath = "https://webapi.teamviewer.com/api/v1"

// WebMonitoringDatasource is an example datasource used to scaffold
// new datasource plugins with an backend.
type WebMonitoringDatasource struct{}

var (
	_ backend.QueryDataHandler      = (*WebMonitoringDatasource)(nil)
	_ backend.CheckHealthHandler    = (*WebMonitoringDatasource)(nil)
	_ instancemgmt.InstanceDisposer = (*WebMonitoringDatasource)(nil)
)

// NewSampleDatasource creates a new datasource instance.
func NewWebMonitoringDatasource(_ backend.DataSourceInstanceSettings) (instancemgmt.Instance, error) {
	return &WebMonitoringDatasource{}, nil
}

type monitor struct {
	MonitorID   string `json:"monitorId"`
	MonitorType string `json:"type"`
	Name        string `json:"name"`
	URL         string `json:"url"`
}

type monitorsResponse struct {
	Monitors          []monitor `json:"monitors"`
	ContinuationToken string    `json:"continuationToken"`
}

func (td *WebMonitoringDatasource) getLocations(ctx context.Context, apiToken string) ([]location, error) {
	queryURL := webMonitingAPIBasePath + "/webMonitoring/locations"

	body, err := doWebMonitoringAPIQuery(ctx, queryURL, apiToken)
	if err != nil {
		log.DefaultLogger.Error(err.Error())

		return nil, errors.New("api call failed")
	}

	log.DefaultLogger.Debug(fmt.Sprintf("Locations (raw): %v", string(body)))

	locations := make([]location, 0)

	err = json.Unmarshal(body, &locations)
	if err != nil {
		log.DefaultLogger.Error("json unmarshall: ", err.Error())

		return nil, errors.New("couldn't parse location json")
	}

	return locations, nil
}

func (td *WebMonitoringDatasource) getMonitors(ctx context.Context, apiToken string) ([]monitor, error) {
	monitors := make([]monitor, 0)

	var continuationToken string

	for {
		u, err := url.Parse(webMonitingAPIBasePath + "/webMonitoring/monitors")
		if err != nil {
			log.DefaultLogger.Error("Couldn't parse API call: ", err.Error())

			return nil, errors.New("couldn't parse API call")
		}

		q := u.Query()

		if continuationToken != "" {
			q.Set("continuationToken", continuationToken)
		}

		u.RawQuery = q.Encode()

		// start request
		body, err := doWebMonitoringAPIQuery(ctx, u.String(), apiToken)
		if err != nil {
			log.DefaultLogger.Error(err.Error())

			return nil, errors.New("get monitors API call failed")
		}

		var resp monitorsResponse

		err = json.Unmarshal(body, &resp)
		if err != nil {
			log.DefaultLogger.Error("json unmarshall: ", err.Error())

			return nil, errors.New("parsing json failed")
		}

		monitors = append(monitors, resp.Monitors...)

		if resp.ContinuationToken != "" {
			continuationToken = resp.ContinuationToken
		} else {
			break
		}
	}

	return monitors, nil
}

func (td *WebMonitoringDatasource) getMonitorResults(ctx context.Context, apiToken, monitorID string,
	timeFrom, timeTo time.Time) ([]monitorResult, error) {
	result := make([]monitorResult, 0)

	var continuationToken string

	// Request monitor results
	for {
		u, err := url.Parse(webMonitingAPIBasePath + "/webMonitoring/monitorResults")
		if err != nil {
			log.DefaultLogger.Error("Couldn't parse API call: ", err.Error())

			return result, errors.New("invalid url")
		}

		q := u.Query()
		q.Set("monitorid", monitorID)
		q.Set("start", timeFrom.Format("2006-01-02T15:04:05Z07:00"))
		q.Set("end", timeTo.Format("2006-01-02T15:04:05Z07:00"))

		if continuationToken != "" {
			q.Set("continuationToken", continuationToken)
		}

		u.RawQuery = q.Encode()

		log.DefaultLogger.Debug(fmt.Sprintf("Requesting monitor results, MonitorID: %v, From: %v, To: %v, ContinuationToken: %v",
			monitorID, timeFrom, timeTo, continuationToken))

		body, err := doWebMonitoringAPIQuery(ctx, u.String(), apiToken)
		if err != nil {
			log.DefaultLogger.Error(err.Error())

			return result, errors.New("api call failed")
		}

		log.DefaultLogger.Debug(fmt.Sprintf("MonitorResults (raw): %v", string(body)))

		var monitorResults monitorResultsResponse

		err = json.Unmarshal(body, &monitorResults)
		if err != nil {
			log.DefaultLogger.Error("json unmarshall: ", err.Error())

			return result, errors.New("parsing response failed")
		}

		log.DefaultLogger.Debug(fmt.Sprintf("Results in total: %v, ContinuationToken: %v",
			len(monitorResults.MonitorResults), monitorResults.ContinuationToken))
		log.DefaultLogger.Debug(fmt.Sprintf("Results: %v", monitorResults.MonitorResults))

		result = append(result, monitorResults.MonitorResults...)

		if monitorResults.ContinuationToken != "" {
			continuationToken = monitorResults.ContinuationToken
		} else {
			break
		}
	}

	return result, nil
}

type alarm struct {
	MonitorID      string    `json:"monitorId"`
	AlarmType      string    `json:"alarmType"`
	FoundAt        time.Time `json:"foundAt"`
	ResolvedAt     time.Time `json:"resolvedAt"`
	AcknowledgedAt time.Time `json:"acknowledgedAt"`
	Duration       string    `json:"duration"`
	Status         string    `json:"alarmStatus"`
}

type alarmResponse struct {
	Alarms            []alarm `json:"alarms"`
	ContinuationToken string  `json:"continuationToken"`
}

func (td *WebMonitoringDatasource) getAlarms(ctx context.Context, apiToken string, timeFrom, timeTo time.Time) ([]alarm, error) {
	alarms := make([]alarm, 0)

	var continuationToken string

	for {
		u, err := url.Parse(webMonitingAPIBasePath + "/webMonitoring/alarms")
		if err != nil {
			log.DefaultLogger.Error("Couldn't parse API call: ", err.Error())

			return nil, errors.New("couldn't parse API call")
		}

		q := u.Query()
		q.Set("start", timeFrom.Format("2006-01-02T15:04:05Z07:00"))
		q.Set("end", timeTo.Format("2006-01-02T15:04:05Z07:00"))

		if continuationToken != "" {
			q.Set("continuationToken", continuationToken)
		}

		log.DefaultLogger.Debug(fmt.Sprintf("Requesting alarms, Start: %v, End: %v, ContiuationToken: %v",
			timeFrom, timeTo, continuationToken))

		u.RawQuery = q.Encode()

		// start request
		body, err := doWebMonitoringAPIQuery(ctx, u.String(), apiToken)
		if err != nil {
			log.DefaultLogger.Error(err.Error())

			return nil, errors.New("get monitors API call failed")
		}

		var resp alarmResponse

		err = json.Unmarshal(body, &resp)
		if err != nil {
			log.DefaultLogger.Error("json unmarshall: ", err.Error())

			return nil, errors.New("parsing json failed")
		}

		log.DefaultLogger.Debug(fmt.Sprintf("Received %v alarms",
			len(resp.Alarms)))

		alarms = append(alarms, resp.Alarms...)

		if resp.ContinuationToken != "" {
			continuationToken = resp.ContinuationToken
		} else {
			break
		}
	}

	return alarms, nil
}

// CallResource is a generic handler to query arbitrary data.
func (td *WebMonitoringDatasource) CallResource(ctx context.Context, req *backend.CallResourceRequest,
	sender backend.CallResourceResponseSender) error {
	log.DefaultLogger.Debug("CallResource", "request", req)

	apiToken := req.PluginContext.DataSourceInstanceSettings.DecryptedSecureJSONData["apiToken"]

	log.DefaultLogger.Debug(fmt.Sprintf("ApiKey: %v", apiToken))

	if apiToken == "" {
		return errors.New("invalid api token")
	}

	response := &backend.CallResourceResponse{}

	log.DefaultLogger.Debug(fmt.Sprintf("Path: %s", req.Path))

	if req.Path == "rm/webmonitoring/monitors" {
		monitors, err := td.getMonitors(ctx, apiToken)
		if err != nil {
			log.DefaultLogger.Error("get monitors failed: ", err.Error())

			return errors.New("get monitors failed")
		}

		b, err := json.Marshal(monitors)
		if err != nil {
			log.DefaultLogger.Error("json marshall: ", err.Error())

			return errors.New("serializing json failed")
		}

		response.Body = b
		response.Status = 200
	} else {
		response.Status = 404
	}

	if err := sender.Send(response); err != nil {
		log.DefaultLogger.Error("send response failed: ", err.Error())

		return errors.New("send response failed")
	}

	return nil
}

// QueryData handles multiple queries and returns multiple responses.
// req contains the queries []DataQuery (where each query contains RefID as a unique identifier).
// The QueryDataResponse contains a map of RefID to the response for each query, and each response
// contains Frames ([]*Frame).
func (td *WebMonitoringDatasource) QueryData(ctx context.Context, req *backend.QueryDataRequest) (*backend.QueryDataResponse, error) {
	log.DefaultLogger.Debug("QueryData", "request", req)

	// create response struct
	response := backend.NewQueryDataResponse()

	apiToken := req.PluginContext.DataSourceInstanceSettings.DecryptedSecureJSONData["apiToken"]

	now := time.Now()

	log.DefaultLogger.Debug(fmt.Sprintf("ApiKey: %v", apiToken))

	if apiToken == "" {
		return response, nil
	}

	// loop over queries and execute them individually.
	for i := range req.Queries {
		res := td.query(ctx, &req.Queries[i], apiToken)

		// save the response in a hashmap
		// based on with RefID as identifier
		response.Responses[req.Queries[i].RefID] = res
	}

	log.DefaultLogger.Debug(fmt.Sprintf("QueryData finished, time: %s", time.Since(now)))

	return response, nil
}

type tokenValid struct {
	Valid bool `json:"token_valid"`
}

type queryModel struct {
	Product   string `json:"queryProduct"`
	Type      string `json:"queryType"`
	MonitorID string `json:"queryMonitorID"`
}

type monitorResultsResponse struct {
	MonitorResults    []monitorResult `json:"monitorResults"`
	ContinuationToken string          `json:"continuationToken"`
}

type monitorResult struct {
	LocationID   int       `json:"locationId"`
	Time         time.Time `json:"time"`
	Status       string    `json:"status"`
	ResponseTime int       `json:"responseTimeMs"`
}

type location struct {
	LocationID  int    `json:"locationId"`
	Continent   string `json:"continent"`
	CountryCode string `json:"countryCode"`
	City        string `json:"city"`
}

func (td *WebMonitoringDatasource) query(ctx context.Context, query *backend.DataQuery, apiToken string) backend.DataResponse {
	// Unmarshal the json into our queryModel
	var qm queryModel

	response := backend.DataResponse{}

	log.DefaultLogger.Debug(fmt.Sprintf("JSON: %v", string(query.JSON)))
	log.DefaultLogger.Debug(fmt.Sprintf("Interval: %v", query.Interval.String()))
	log.DefaultLogger.Debug(fmt.Sprintf("MaxDataPoints: %v", query.MaxDataPoints))
	log.DefaultLogger.Debug(fmt.Sprintf("TimeRange: [%v,%v]", query.TimeRange.From.String(), query.TimeRange.To.String()))

	response.Error = json.Unmarshal(query.JSON, &qm)
	if response.Error != nil {
		return response
	}

	if qm.Product != "webmonitoring" {
		response.Error = fmt.Errorf("invalid product: '%s'", qm.Product)

		return response
	}

	switch {
	case qm.Type == "monitorresults":
		// Log a warning if `MonitorID` is empty.
		if qm.MonitorID == "" {
			log.DefaultLogger.Error("MonitorID is empty")

			response.Error = errors.New("invalid monitor id")

			return response
		}

		log.DefaultLogger.Info(fmt.Sprintf("MonitorID: %v", qm.MonitorID))

		// Request locations
		locations, err := td.getLocations(ctx, apiToken)
		if err != nil {
			log.DefaultLogger.Error("getLocations: ", err.Error())

			response.Error = errors.New("get locations failed")

			return response
		}

		// Get monitor results
		monitorResults, err := td.getMonitorResults(ctx, apiToken, qm.MonitorID, query.TimeRange.From, query.TimeRange.To)
		if err != nil {
			log.DefaultLogger.Error("getMonitorResults: ", err.Error())

			response.Error = errors.New("get monitor results failed")

			return response
		}

		type LocationResults struct {
			times  []time.Time
			values []int32
		}

		resultMap := make(map[int]LocationResults)

		for _, mr := range monitorResults {
			tmp := resultMap[mr.LocationID]

			tmp.times = append(tmp.times, mr.Time)
			tmp.values = append(tmp.values, int32(mr.ResponseTime))

			resultMap[mr.LocationID] = tmp
		}

		locationMap := make(map[int]string)
		locationMapReverse := make(map[string]int)

		for i := 0; i < len(locations); i++ {
			locationID := locations[i].LocationID
			locationName := locations[i].City + " (" + strings.ToUpper(locations[i].CountryCode) + ")"
			locationMap[locationID] = locationName
			locationMapReverse[locationName] = locationID
		}

		locationNames := make([]string, 0)
		for k := range resultMap {
			locationNames = append(locationNames, locationMap[k])
		}

		sort.Strings(locationNames)

		for _, locationName := range locationNames {
			locationID := locationMapReverse[locationName]

			_, ok := resultMap[locationID]
			if !ok {
				continue
			}

			log.DefaultLogger.Debug(fmt.Sprintf("LocationID: %v, LocationName: %v, %v entries",
				locationID, locationName, len(resultMap[locationID].times)))

			log.DefaultLogger.Debug(fmt.Sprintf("Times: %v entries, %v",
				len(resultMap[locationID].times), resultMap[locationID].times))
			log.DefaultLogger.Debug(fmt.Sprintf("Values: %v entries, %v",
				len(resultMap[locationID].values), resultMap[locationID].values))

			// create data frame response
			frame := data.NewFrame("response")

			frame.Fields = append(frame.Fields,
				data.NewField("time", nil, resultMap[locationID].times),        // time dimension
				data.NewField(locationName, nil, resultMap[locationID].values), // values
			)

			config := &data.FieldConfig{}
			config.Unit = "ms"

			frame.Fields[1].SetConfig(config)

			// add the frames to the response
			response.Frames = append(response.Frames, frame)
		}
	case qm.Type == "alarms":
		monitors, err := td.getMonitors(ctx, apiToken)
		if err != nil {
			log.DefaultLogger.Error("get monitors failed: ", err.Error())

			response.Error = errors.New("get monitors failed")

			return response
		}

		monitorsMap := make(map[string]string)

		for _, m := range monitors {
			monitorsMap[m.MonitorID] = m.Name
		}

		// Request alarms
		alarms, err := td.getAlarms(ctx, apiToken, query.TimeRange.From.UTC(), query.TimeRange.To.UTC())
		if err != nil {
			log.DefaultLogger.Error("getAlarms: ", err.Error())

			response.Error = errors.New("get alarms failed")

			return response
		}

		log.DefaultLogger.Debug(fmt.Sprintf("Received %v alarms in total",
			len(alarms)))

		var monitorNames, alarmStatus, alarmTypes, foundAt, resolvedAt, acknowledgedAt, duration []string

		location, err := time.LoadLocation("UTC")
		if err != nil {
			log.DefaultLogger.Error("get location UTC: ", err.Error())
		}

		for idx := range alarms {
			m, ok := monitorsMap[alarms[idx].MonitorID]
			if !ok {
				continue
			}

			monitorNames = append(monitorNames, m)
			alarmStatus = append(alarmStatus, alarms[idx].Status)
			alarmTypes = append(alarmTypes, alarms[idx].AlarmType)
			foundAt = append(foundAt, alarms[idx].FoundAt.In(location).Format(time.RFC3339Nano))

			if alarms[idx].ResolvedAt.IsZero() {
				resolvedAt = append(resolvedAt, "")
			} else {
				resolvedAt = append(resolvedAt, alarms[idx].ResolvedAt.In(location).Format(time.RFC3339Nano))
			}

			if alarms[idx].AcknowledgedAt.IsZero() {
				acknowledgedAt = append(acknowledgedAt, "")
			} else {
				acknowledgedAt = append(acknowledgedAt, alarms[idx].AcknowledgedAt.In(location).Format(time.RFC3339Nano))
			}

			duration = append(duration, alarms[idx].Duration)
		}

		log.DefaultLogger.Debug(fmt.Sprintf("MonitorIDs: %v entries, %v",
			len(monitorNames), monitorNames))
		log.DefaultLogger.Debug(fmt.Sprintf("AlarmStatus: %v entries, %v",
			len(alarmStatus), alarmStatus))
		log.DefaultLogger.Debug(fmt.Sprintf("AlarmType: %v entries, %v",
			len(alarmTypes), alarmTypes))

		// create data frame response
		frame := data.NewFrame("response")

		frame.Fields = append(frame.Fields,
			data.NewField("Monitor ID", nil, monitorNames),
			data.NewField("Alarm Type", nil, alarmTypes),
			data.NewField("Status", nil, alarmStatus),
			data.NewField("Found", nil, foundAt),
			data.NewField("Resolved", nil, resolvedAt),
			data.NewField("Acknowledged", nil, acknowledgedAt),
			data.NewField("Duration", nil, duration))

		// add the frames to the response
		response.Frames = append(response.Frames, frame)
	case qm.Type == "monitors":
		monitors, err := td.getMonitors(ctx, apiToken)
		if err != nil {
			log.DefaultLogger.Error("get monitors failed: ", err.Error())

			response.Error = errors.New("get monitors failed")

			return response
		}

		var monitorNames, monitorTypes, monitorURLs []string

		for _, monitor := range monitors {
			monitorNames = append(monitorNames, monitor.Name)
			monitorTypes = append(monitorTypes, monitor.MonitorType)
			monitorURLs = append(monitorURLs, monitor.URL)
		}

		// create data frame response
		frame := data.NewFrame("response")

		frame.Fields = append(frame.Fields,
			data.NewField("Name", nil, monitorNames),
			data.NewField("Monitor Type", nil, monitorTypes),
			data.NewField("URL", nil, monitorURLs))

		// add the frames to the response
		response.Frames = append(response.Frames, frame)
	default:
		response.Error = fmt.Errorf("invalid query Type: '%s'", qm.Type)
	}

	return response
}

// CheckHealth handles health checks sent from Grafana to the plugin.
// The main use case for these health checks is the test button on the
// datasource configuration page which allows users to verify that
// a datasource is working as expected.
func (td *WebMonitoringDatasource) CheckHealth(ctx context.Context, req *backend.CheckHealthRequest) (*backend.CheckHealthResult, error) {
	status, message := checkAPIToken(ctx, req.PluginContext.DataSourceInstanceSettings.DecryptedSecureJSONData["apiToken"])

	return &backend.CheckHealthResult{
		Status:  status,
		Message: message,
	}, nil
}

// Dispose here tells plugin SDK that plugin wants to clean up resources when a new instance
// created. As soon as datasource settings change detected by SDK old datasource instance will
// be disposed and a new one will be created using NewSampleDatasource factory function.
func (td *WebMonitoringDatasource) Dispose() {
	// Clean up datasource instance resources.
}

// checkAPIToken do a API call to /ping for checking if the token is valid.
func checkAPIToken(ctx context.Context, token string) (status backend.HealthStatus, message string) {
	const couldntCheck = "Couldn't check Token validity"

	var tokenStatus tokenValid

	log.DefaultLogger.Info("Configuring datasource plugin")

	queryURL := webMonitingAPIBasePath + "/ping"

	body, err := doWebMonitoringAPIQuery(ctx, queryURL, token)
	if err != nil {
		log.DefaultLogger.Error(err.Error())

		return backend.HealthStatusError, err.Error()
	}

	err = json.Unmarshal(body, &tokenStatus)
	if err != nil {
		log.DefaultLogger.Error("json unmarshall: ", err.Error(), body)

		return backend.HealthStatusUnknown, couldntCheck
	}

	if !tokenStatus.Valid {
		return backend.HealthStatusError, "Invalid Token"
	} else if tokenStatus.Valid {
		return backend.HealthStatusOk, "Data source is working"
	}

	return backend.HealthStatusUnknown, couldntCheck
}

func doWebMonitoringAPIQuery(ctx context.Context, queryURL, apiToken string) (body []byte, err error) {
	client := &http.Client{}

	log.DefaultLogger.Debug(fmt.Sprintf("Starting request %s", queryURL))

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, queryURL, nil)
	if err != nil {
		log.DefaultLogger.Warn("HTTP New request: ", err.Error())

		return body, fmt.Errorf("HTTP New request: %w", err)
	}

	req.Header.Add("Authorization", "Bearer "+apiToken)

	now := time.Now()

	res, err := client.Do(req)
	if res.StatusCode != http.StatusOK {
		log.DefaultLogger.Warn(fmt.Sprintf("HTTP request returned %s", res.Status))

		return body, fmt.Errorf("HTTP request returned %s", res.Status)
	} else if err != nil {
		log.DefaultLogger.Warn(fmt.Sprintf("HTTP request do: %s", err.Error()))

		return body, fmt.Errorf("HTTP request do: %w", err)
	}
	defer res.Body.Close()

	body, err = ioutil.ReadAll(res.Body)
	if err != nil {
		log.DefaultLogger.Warn(fmt.Sprintf("HTTP readall: %s", err.Error()))

		return body, fmt.Errorf("HTTP readall: %w", err)
	}

	elapsed := time.Since(now)

	log.DefaultLogger.Debug(fmt.Sprintf("Request finished, time: %s", elapsed))
	log.DefaultLogger.Debug(string(body))

	return body, nil
}
