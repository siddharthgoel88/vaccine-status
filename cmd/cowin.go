package cmd

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"github.com/prometheus/common/log"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

const (
	COWIN_URL = "https://cdn-api.co-vin.in"
	BY_DISTRICT_API = "api/v2/appointment/sessions/public/calendarByDistrict"
)

func GetSlotsByDistrict(districtId int){
	log.Infof("Checking for slots %d\n", districtId)

	loc, _ := time.LoadLocation("Asia/Kolkata")
	now := time.Now().In(loc)
	firstWeek := now.Format("02-01-2006")
	calendarByDistrictFirstWeek, err := getCalendarByDistrictHttp(districtId, firstWeek)
	if err != nil {
		log.Errorf("error in http request, err=%v", err)
		return
	}

	secondWeek := now.Add(time.Hour * 24 * 7).In(loc).Format("02-01-2006")
	calendarByDistrictSecondWeek, err := getCalendarByDistrictHttp(districtId, secondWeek)
	if err != nil {
		log.Errorf("error in http request, err=%v", err)
		return
	}

	result := &Result{
		dose1over18: make([]*SlotsAvailability, 0),
		dose2over18: make([]*SlotsAvailability, 0),
		dose1over45: make([]*SlotsAvailability, 0),
		dose2over45: make([]*SlotsAvailability, 0),
	}
	filterAvailableSlots(calendarByDistrictFirstWeek, result)
	filterAvailableSlots(calendarByDistrictSecondWeek, result)

	createResultFile(result, districtId)
}

func createResultFile(result *Result, districtId int) {
	// create directory if does not exists
	timeInSec := time.Now().Unix()

	if len(result.dose1over18) > 0 {
		createResultFilePerFilter(1, 18, districtId, timeInSec, result.dose1over18)
	}

	if len(result.dose2over18) > 0 {
		createResultFilePerFilter(2, 18, districtId, timeInSec, result.dose2over18)
	}

	if len(result.dose1over45) > 0 {
		createResultFilePerFilter(1, 45, districtId, timeInSec, result.dose1over45)
	}

	if len(result.dose2over45) > 0 {
		createResultFilePerFilter(2, 45, districtId, timeInSec, result.dose2over45)
	}

}

func createResultFilePerFilter(doseNumber, age, districtId int, timeInSec int64, slots []*SlotsAvailability) {
	if len(slots) == 0 {
		return
	}

	path := fmt.Sprintf("./results/%d/%d/dose-%d", districtId, age, doseNumber)

	resultPath := filepath.FromSlash(path)
	_ = os.MkdirAll(resultPath, os.ModePerm)

	outputFile, err := os.Create(fmt.Sprintf("%s/%d.csv", resultPath, timeInSec))
	if err != nil {
		log.Errorf("error in creating file, err=%v\n", err)
		return
	}
	defer func(){ _ = outputFile.Close() }()

	values := slotAvailabilityToArray(slots)

	w := csv.NewWriter(outputFile)
	if err := w.WriteAll(values); err != nil {
		log.Errorf("error in writing csv  err=%v\n", err)
		return
	}
	log.Debugf("successfully written csv file=%v", outputFile.Name())
}

func filterAvailableSlots(calendarByDistrictFirstWeek *CalendarByDistrictResponse, result *Result) {
	for _, centre := range calendarByDistrictFirstWeek.Centers {
		for _, session := range centre.Sessions {
			if session.AvailableCapacityDose1 > 0 || session.AvailableCapacityDose2 > 0 {
				
				slotDose1 := &SlotsAvailability{
					centreName:   centre.Name,
					districtName: centre.DistrictName,
					pincode:      fmt.Sprintf("%d",centre.Pincode),
					sessionDate:  session.Date,
					availability: session.AvailableCapacityDose1,
					minAgeLimit:  session.MinAgeLimit,
					vaccine:      session.Vaccine,
				}

				slotDose2 := &SlotsAvailability{
					centreName:   centre.Name,
					districtName: centre.DistrictName,
					pincode:      fmt.Sprintf("%d",centre.Pincode),
					sessionDate:  session.Date,
					availability: session.AvailableCapacityDose2,
					minAgeLimit:  session.MinAgeLimit,
					vaccine:      session.Vaccine,
				}

				if session.MinAgeLimit == 18 && session.AvailableCapacityDose1 > 0{
					result.dose1over18 = append(result.dose1over18, slotDose1)
				}

				if session.MinAgeLimit == 18 && session.AvailableCapacityDose2 > 0{
					result.dose2over18 = append(result.dose2over18, slotDose2)
				}

				if session.MinAgeLimit == 45 && session.AvailableCapacityDose1 > 0{
					result.dose1over45 = append(result.dose1over45, slotDose1)
				}

				if session.MinAgeLimit == 45 && session.AvailableCapacityDose2 > 0{
					result.dose2over45 = append(result.dose2over45, slotDose2)
				}
			}
		}
	}
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

func slotAvailabilityToArray(slots []*SlotsAvailability) [][]string {
	slotsArray := make([][]string, 0)

	header := []string{"Center Name", "District", "Pincode", "Session Date", "Availability", "Min Age Limit", "Vaccine"}
	slotsArray = append(slotsArray, header)

	for _, slot := range slots {
		row := []string{slot.centreName, slot.districtName, slot.pincode, slot.sessionDate,
			fmt.Sprintf("%d",int(slot.availability)), fmt.Sprintf("%d", slot.minAgeLimit), slot.vaccine}
		slotsArray = append(slotsArray, row)
	}

	return slotsArray
}

type Result struct {
	dose1over18 []*SlotsAvailability
	dose2over18 []*SlotsAvailability
	dose1over45 []*SlotsAvailability
	dose2over45 []*SlotsAvailability
}

type SlotsAvailability struct {
	centreName string
	districtName string
	pincode string
	sessionDate string
	availability float64
	minAgeLimit int
	vaccine string
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
			SessionID              string   `json:"session_id"`
			Date                   string   `json:"date"`
			AvailableCapacity      float64  `json:"available_capacity"`
			MinAgeLimit            int      `json:"min_age_limit"`
			Vaccine                string   `json:"vaccine"`
			Slots                  []string `json:"slots"`
			AvailableCapacityDose1 float64  `json:"available_capacity_dose1"`
			AvailableCapacityDose2 float64  `json:"available_capacity_dose2"`
		} `json:"sessions"`
	} `json:"centers"`
}
