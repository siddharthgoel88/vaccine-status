package cmd

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"github.com/prometheus/common/log"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"time"
)

const (
	COWIN_URL = "https://cdn-api.co-vin.in"
	BY_DISTRICT_API = "api/v2/appointment/sessions/public/calendarByDistrict"
)

func GetSlotsByDistrict(districtId int) (*os.File, error) {
	loc, _ := time.LoadLocation("Asia/Kolkata")
	now := time.Now().In(loc)
	firstWeek := now.Format("02-01-2006")
	calendarByDistrictFirstWeek, err := getCalendarByDistrictHttp(districtId, firstWeek)
	if err != nil {
		return nil, err
	}

	secondWeek := now.Add(time.Hour * 24 * 7).In(loc).Format("02-01-2006")
	calendarByDistrictSecondWeek, err := getCalendarByDistrictHttp(districtId, secondWeek)
	if err != nil {
		return nil, err
	}

	values := make([][]string, 0)
	header := []string{"Center Name", "District", "Pincode", "Session Date", "Availability", "Min Age Limit", "Vaccine"}
	values = append(values, header)

	for _, centre := range calendarByDistrictFirstWeek.Centers {
		for _, session := range centre.Sessions {
			if session.AvailableCapacity != 0 && session.MinAgeLimit == 18 {
				value := []string{centre.Name, centre.DistrictName, strconv.Itoa(centre.Pincode),
					session.Date, fmt.Sprintf("%f",session.AvailableCapacity),
					strconv.Itoa(session.MinAgeLimit), session.Vaccine}
				values = append(values, value)
			}
		}
	}

	for _, centre := range calendarByDistrictSecondWeek.Centers {
		for _, session := range centre.Sessions {
			if session.AvailableCapacity != 0 && session.MinAgeLimit == 18 {
				value := []string{centre.Name, centre.DistrictName, strconv.Itoa(centre.Pincode),
					session.Date, fmt.Sprintf("%f",session.AvailableCapacity),
					strconv.Itoa(session.MinAgeLimit), session.Vaccine}
				values = append(values, value)
			}
		}
	}

	if len(values) == 1 {
		return nil, nil
	}

	outputFile, err := os.Create("result.csv")
	if err != nil {
		// handle error
	}
	defer outputFile.Close()
	w := csv.NewWriter(outputFile)
	if err := w.WriteAll(values); err != nil {
		log.Error("error in writing csv file")
		return nil, err
	}

	return outputFile, nil
}

func getCalendarByDistrictHttp(districtId int, startDate string) (*CalendarByDistrictResponse, error) {

	// construct HTTP request
	payload := fmt.Sprintf("%s/%s?district_id=%d&date=%s", COWIN_URL, BY_DISTRICT_API, districtId, startDate)
	req, err := http.NewRequest("GET", payload, nil)
	if err != nil {
		return nil, err
		// handle err
	}
	req.Header.Set("Authority", "cdn-api.co-vin.in")
	req.Header.Set("Pragma", "no-cache")
	req.Header.Set("Cache-Control", "no-cache")
	req.Header.Set("Sec-Ch-Ua", "\" Not A;Brand\";v=\"99\", \"Chromium\";v=\"90\", \"Google Chrome\";v=\"90\"")
	req.Header.Set("Sec-Ch-Ua-Mobile", "?0")
	req.Header.Set("Upgrade-Insecure-Requests", "1")
	req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/90.0.4430.93 Safari/537.36")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		err = fmt.Errorf("request did not return 200. resp=%v", resp.StatusCode)
		return &CalendarByDistrictResponse{}, err
	}

	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var calendarByDistrictResponse CalendarByDistrictResponse
	err = json.Unmarshal(bodyBytes, &calendarByDistrictResponse)
	if err != nil {
		return nil, err
	}

	return &calendarByDistrictResponse, nil
}

type CalendarByDistrictResponse struct {
	Centers []struct {
		CenterID     int    `json:"center_id"`
		Name         string `json:"name"`
		Address      string `json:"address"`
		StateName    string `json:"state_name"`
		DistrictName string `json:"district_name"`
		BlockName    string `json:"block_name"`
		Pincode      int    `json:"pincode"`
		Lat          int    `json:"lat"`
		Long         int    `json:"long"`
		From         string `json:"from"`
		To           string `json:"to"`
		FeeType      string `json:"fee_type"`
		Sessions     []struct {
			SessionID         string   `json:"session_id"`
			Date              string   `json:"date"`
			AvailableCapacity float64  `json:"available_capacity"`
			MinAgeLimit       int      `json:"min_age_limit"`
			Vaccine           string   `json:"vaccine"`
			Slots             []string `json:"slots"`
		} `json:"sessions"`
	} `json:"centers"`
}
