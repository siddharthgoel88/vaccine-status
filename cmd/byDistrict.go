/*
Copyright © 2021 Siddharth Goel siddharth98391@gmail.com

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package cmd

import (
	"bufio"
	"encoding/base64"
	"fmt"
	"github.com/prometheus/common/log"
	"github.com/sendgrid/sendgrid-go"
	"github.com/sendgrid/sendgrid-go/helpers/mail"
	"github.com/spf13/cobra"
	"io/ioutil"
	"os"
	"strconv"
	"strings"
	"time"
)

// byDistrictCmd represents the byDistrict command
var byDistrictCmd = &cobra.Command{
	Use:   "byDistrict",
	Short: "Give the district id here",
	Run: func(cmd *cobra.Command, args []string) {
		// checking the number of arguments
		if len(args) == 0 {
			log.Error("pass one or more district id(s) as argument")
			return
		}

		districtIds := make([]int, 0)

		for _,arg := range args {
			// checking district id is integer
			districtId, err := strconv.Atoi(arg);
			if err != nil {
				log.Error("district [%v] id must be an integer", arg)
				return
			}

			districtIds = append(districtIds, districtId)
		}

		alertedIds := make(map[int]time.Time)

		for true {

			if len(alertedIds) == len(districtIds) {
				break
			}

			for _, districtId := range districtIds {
				if startTime, ok := alertedIds[districtId]; ok {
					if time.Now().Sub(startTime).Seconds() < 60 {
						log.Infof("skipping alerts for district %d", districtId)
						continue
					}
					delete(alertedIds, districtId)
				}

				log.Info("sleeping for 5 seconds before next call")
				time.Sleep(5 * time.Second)

				file, err := GetSlotsByDistrict(districtId)
				if err != nil {
					log.Errorf("issue in HTTP request, err = %v", err)
					continue
				}

				if file == nil {
					log.Infof("No available slot found for district %d", districtId)
					continue
				}

				log.Infof("Found available slots for district %d", districtId)
				alertedIds[districtId] = time.Now()

				// alerting via Email
				from := mail.NewEmail("Vaccine Baba", "no-reply@vaccine-baba.com")
				to := mail.NewEmail("Vaccine Kaka", "kaka@vaccine-baba.com")
				tos := []*mail.Email{to}
				subject := fmt.Sprintf("Vaccination slot for District %d", districtId)
				htmlContent := fmt.Sprintf("Hey buddy,<br><br>You are getting this email " +
					"since you have subscribed to vaccination alert in district %d. This email " +
					"contains list of available slots at one or more of the subscribed " +
					"locations. Check the mail attachment. <br> Stay safe, stay vaccinated." +
					"<br><br>Cheers,<br>Vaccine Baba", districtId)
				content := mail.NewContent("text/html", htmlContent)

				attachment := mail.NewAttachment()
				attachment.SetType("text/csv")
				attachment.SetFilename(file.Name())
				b64, err := getBase64EncodedData(file.Name())
				if err != nil {
					log.Errorf("%v", err)
					continue
				}
				attachment.SetContent(b64)

				m := mail.NewV3MailInit(from, subject, to, content)
				m.AddAttachment(attachment)

				p := mail.NewPersonalization()
				p.AddTos(tos...)
				p.AddBCCs(getEmailList(districtId)...)
				m.AddPersonalizations(p)

				client := sendgrid.NewSendClient(os.Getenv("SENDGRID_API_KEY"))
				response, err := client.Send(m)
				if err != nil {
					log.Errorf("%v", err)
				} else {
					fmt.Println(response.StatusCode)
					fmt.Println(response.Body)
					fmt.Println(response.Headers)
				}
			}
		}
	},
}

func getBase64EncodedData(fileName string) (string,error) {
	file, err := os.Open(fileName)
	if err != nil {
		return "", err
	}

	// Read entire JPG into byte slice.
	reader := bufio.NewReader(file)
	data, err := ioutil.ReadAll(reader)

	return base64.StdEncoding.EncodeToString(data), nil
}

func getEmailList(districtId int) []*mail.Email {
	emails, exists := os.LookupEnv(fmt.Sprintf("EMAIL_%d", districtId))
	if !exists {
		log.Errorf("no value set in EMAIL_%d", districtId)
		return []*mail.Email{}
	}

	emailList := strings.Fields(emails)
	output := make([]*mail.Email, 0)

	for _,email := range emailList {
		log.Info(email)
		emailObj := &mail.Email{Address: email}
		output = append(output, emailObj)
	}

	return output
}

func fileExists(filename string) bool {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}

func init() {
	rootCmd.AddCommand(byDistrictCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// byDistrictCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// byDistrictCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
