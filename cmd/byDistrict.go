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
	"fmt"
	"github.com/prometheus/common/log"
	"github.com/spf13/cobra"
	"strconv"
	"time"
)

const (
	MAX_RUNTIME = 82000
	SLEEP_SECONDS = 15
)


// byDistrictCmd represents the byDistrict command
var byDistrictCmd = &cobra.Command{
	Use:   "byDistrict",
	Short: "Give the district id here",
	Run: func(cmd *cobra.Command, args []string) {
		var districtId int

		// checking the number of arguments
		if len(args) != 1 {
			log.Error("pass district id as an argument")
			return
		}

		// checking district id is integer
		districtId, err := strconv.Atoi(args[0]);
		if err != nil {
			log.Error("district id must be an integer")
			return
		}

		runTime := 0

		for runTime < MAX_RUNTIME {
			file, err := GetSlotsByDistrict(districtId)
			if err != nil {
				log.Errorf("issue in HTTP request, err = %v", err)
			}

			if file != nil {
				fmt.Printf("%s", file.Name())
				break
			}

			time.Sleep(SLEEP_SECONDS * time.Second)
			runTime += SLEEP_SECONDS
		}

	},
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
