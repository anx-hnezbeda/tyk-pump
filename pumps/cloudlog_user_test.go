package pumps

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/TykTechnologies/tyk-pump/analytics"
)

func CreateCloudLogUserRecord(path string, tags []string) analytics.AnalyticsRecord {
	a := analytics.AnalyticsRecord{}
	a.Method = "POST"
	a.Path = path
	a.ContentLength = 123
	a.UserAgent = "Test User Agent"
	a.Day = 26
	a.Month = time.January
	a.Year = 2020
	a.Hour = 9
	a.ResponseCode = 202
	a.APIKey = "APIKEY123"
	a.TimeStamp = time.Now()
	a.APIVersion = "1"
	a.APIName = "Test API"
	a.APIID = "API123"
	a.OrgID = "ORG123"
	a.OauthID = "Oauth123"
	a.RequestTime = time.Now().Unix()
	a.RawRequest = "{\"field\": \"value\"}"
	a.RawResponse = "{\"id\": \"123\"}"
	a.IPAddress = "127.0.0.1"
	a.Tags = tags
	a.ExpireAt = time.Date(2020, time.November, 10, 23, 0, 0, 0, time.UTC)

	return a
}

func TestCloudLogUserPump(t *testing.T) {
	t.Skip("Set the tCloudLogUrl, tCloudToken and remove Skip to test.")

	tCloudLogUrl := "XXX"
	tCloudToken := "XXX"

	tConf := make(map[string]string)
	tConf["url"] = tCloudLogUrl
	tConf["token"] = tCloudToken
	tConf["environment"] = "Testing"

	s := CloudLogUserPump{}

	err := s.Init(tConf)
	if err != nil {
		t.Error(err)
	}

	tData := make([]interface{}, 2)
	tData[0] = CreateCloudLogRecord("/path1", []string{"tag-1", "tag-2"})
	tData[1] = CreateCloudLogRecord("/path2", []string{"tag-1", "tag-2", fmt.Sprintf("cloudlog::%s::%s", tCloudLogUrl, tCloudToken)})

	go s.WriteData(context.TODO(), tData)

	time.Sleep(time.Second)
}