package pumps

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/TykTechnologies/logrus"
	"github.com/TykTechnologies/tyk-pump/analytics"
	"github.com/mitchellh/mapstructure"
	"net/http"
	"strings"
	"time"
)

var cloudLogPumpPrefix = "cloudlog-pump"

type CloudLogPumpConfig struct {
	URL            string `mapstructure:"url"`
	Token          string `mapstructure:"token"`
	Environment    string `mapstructure:"environment"`
}

type CloudLogPump struct {
	clConf  *CloudLogPumpConfig
	timeout int
}

func CloudLogPushData(data []byte, clUrl string, clToken string, prefix string) error {
	req, err := http.NewRequest("POST", clUrl, bytes.NewBuffer(data))
	if err != nil {
		log.WithFields(logrus.Fields{
			"prefix": prefix,
		}).Error("Cannot create new request.", err.Error())

		return err
	}
	req.Header.Set("Authorization", clToken)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.WithFields(logrus.Fields{
			"prefix": prefix,
		}).Error("Cannot post data.", err.Error())

		return err
	}
	defer resp.Body.Close()

	log.WithFields(logrus.Fields{
		"prefix": prefix,
	}).Info("CloudLog request responded with a ", resp.StatusCode, " status code")

	return nil
}

func (p *CloudLogPump) New() Pump {
	return &CloudLogPump{}
}

func (p *CloudLogPump) GetName() string {
	return "CloudLog Pump"
}

func (p *CloudLogPump) Init(conf interface{}) error {
	p.clConf = &CloudLogPumpConfig{}
	err := mapstructure.Decode(conf, p.clConf)
	if err != nil {
		log.WithFields(logrus.Fields{
			"prefix": cloudLogPumpPrefix,
		}).Fatalf("Failed to decode configuration: %s", err)
	}

	log.WithFields(logrus.Fields{
		"prefix": cloudLogPumpPrefix,
	}).Infof("Initializing CloudLog Pump")

	return nil
}

func (p *CloudLogPump) WriteData(ctx context.Context, data []interface{}) error {
	log.WithFields(logrus.Fields{
		"prefix": cloudLogPumpPrefix,
	}).Info("Writing ", len(data), " records")

	var mapping = make(map[string][]map[string]interface{})
	for _, v := range data {
		decoded := v.(analytics.AnalyticsRecord)
		mappedItem := map[string]interface{}{
			"timestamp":       decoded.TimeStamp.Format(time.RFC3339),
			"environment":     p.clConf.Environment,
			"method":          decoded.Method,
			"host":            decoded.Host,
			"path":            decoded.Path,
			"raw_path":        decoded.RawPath,
			"response_code":   decoded.ResponseCode,
			"api_key":         decoded.APIKey,
			"api_version":     decoded.APIVersion,
			"api_name":        decoded.APIName,
			"api_id":          decoded.APIID,
			"org_id":          decoded.OrgID,
			"oauth_id":        decoded.OauthID,
			"raw_request":     decoded.RawRequest,
			"raw_response":    decoded.RawResponse,
			"request_time":    decoded.RequestTime,
			"ip_address":      decoded.IPAddress,
			"user_agent":      decoded.UserAgent,
			// Optional
			"track_path":      decoded.TrackPath,
			"expire_at":       decoded.ExpireAt.Format(time.RFC3339),
			"day":             decoded.Day,
			"month":           decoded.Month,
			"year":            decoded.Year,
			"hour":            decoded.Hour,
			"content_length":  decoded.ContentLength,
			"tags":            decoded.Tags,
			//Geo           GeoData
			//Network       NetworkStats
			//Latency       Latency
			//Alias         string
		}
		p.addCloudLogKeys(decoded.Tags, mappedItem)
		mapping["records"] = append(mapping["records"], mappedItem)
	}

	event, err := json.Marshal(mapping)
	if err != nil {
		return fmt.Errorf("failed to marshal decoded data: %s", err)
	}

	if CloudLogPushData(event, p.clConf.URL, p.clConf.Token, cloudLogPumpPrefix) != nil {
		log.WithFields(logrus.Fields{
			"prefix": cloudLogPumpPrefix,
		}).Error("Cannot log data to cloudlog.")
	}

	return nil
}

func (p *CloudLogPump) SetTimeout(timeout int) {
	p.timeout = timeout
}

func (p *CloudLogPump) GetTimeout() int {
	return p.timeout
}

func (p *CloudLogPump) addCloudLogKeys(tags []string, mappedItem map[string]interface{}) {
	for _, s := range tags {
		conf := strings.Split(s, "::")
		if len(conf) == 3 && conf[0] == "engine-cloudlog" {
			mappedItem[conf[1]] = conf[2]
		}
	}
}
