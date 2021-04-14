package pumps

import (
	"context"
	"encoding/json"
	"github.com/TykTechnologies/logrus"
	"github.com/TykTechnologies/tyk-pump/analytics"
	"github.com/mitchellh/mapstructure"
	"strings"
	"time"
)

var cloudLogUserPumpPrefix = "cloudloguser-pump"

type CloudLogUserPumpConfig struct {
	Environment    string `mapstructure:"environment"`
}

type CloudLogUserPump struct {
	clConf  *CloudLogUserPumpConfig
	timeout int
	CommonPumpConfig
}

func (p *CloudLogUserPump) New() Pump {
	return &CloudLogUserPump{}
}

func (p *CloudLogUserPump) GetName() string {
	return "CloudLog Pump"
}

func (p *CloudLogUserPump) Init(conf interface{}) error {
	p.clConf = &CloudLogUserPumpConfig{}
	err := mapstructure.Decode(conf, p.clConf)
	if err != nil {
		log.WithFields(logrus.Fields{
			"prefix": cloudLogUserPumpPrefix,
		}).Fatalf("Failed to decode configuration: %s", err)
	}

	log.WithFields(logrus.Fields{
		"prefix": cloudLogUserPumpPrefix,
	}).Infof("Initializing CloudLog User Pump")

	return nil
}

func (p *CloudLogUserPump) LogUserData(record analytics.AnalyticsRecord, mappedRecord map[string]interface{}) bool {
	for _, s := range record.Tags {
		conf := strings.Split(s, "::")
		if len(conf) == 3 && conf[0] == "cloudlog" {
			event, err := json.Marshal(mappedRecord)
			if err != nil {
				log.WithFields(logrus.Fields{
					"prefix": cloudLogUserPumpPrefix,
				}).Error("Failed to marshal decoded user data")

				return false
			}

			if CloudLogPushData(event, conf[1], conf[2], cloudLogUserPumpPrefix) != nil {
				log.WithFields(logrus.Fields{
					"prefix": cloudLogUserPumpPrefix,
				}).Error("Failed to log user data to cloudlog.")
			} else {
				return true
			}
		}
	}

	return false
}

func (p *CloudLogUserPump) WriteData(ctx context.Context, data []interface{}) error {
	log.WithFields(logrus.Fields{
		"prefix": cloudLogUserPumpPrefix,
	}).Info("Received ", len(data), " records")

	userRecordCount := 0
	for _, v := range data {
		decoded := v.(analytics.AnalyticsRecord)
		mappedItem := map[string]interface{}{
			"timestamp":       decoded.TimeStamp.Format(time.RFC3339),
			"environment":     p.clConf.Environment,
			"method":          decoded.Method,
			"host":            decoded.Host,
			"response_code":   decoded.ResponseCode,
			"api_key":         decoded.APIKey,
			"api_version":     decoded.APIVersion,
			"api_name":        decoded.APIName,
			"org_id":          decoded.OrgID,
			"oauth_id":        decoded.OauthID,
			"request_time":    decoded.RequestTime,
			"ip_address":      decoded.IPAddress,
			"user_agent":      decoded.UserAgent,
			"track_path":      decoded.TrackPath,
			"expire_at":       decoded.ExpireAt.Format(time.RFC3339),
			"day":             decoded.Day,
			"month":           decoded.Month,
			"year":            decoded.Year,
			"hour":            decoded.Hour,
			"content_length":  decoded.ContentLength,
		}

		// Try to log record as user record
		if p.LogUserData(decoded, mappedItem) == true {
			userRecordCount++
		}
	}

	log.WithFields(logrus.Fields{
		"prefix": cloudLogUserPumpPrefix,
	}).Info("Wrote ", userRecordCount, " records")

	return nil
}

func (p *CloudLogUserPump) SetTimeout(timeout int) {
	p.timeout = timeout
}

func (p *CloudLogUserPump) GetTimeout() int {
	return p.timeout
}