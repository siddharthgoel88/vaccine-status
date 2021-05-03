#!/bin/bash

pincode_file='lucknow-pincode.txt'
START_DATE=03-05-2021
# MIN_AGE=18
MIN_AGE=45

for i in {0..13}
do
    DAYS="${i}d"
    NEXT_DATE=$(date -j -v +$DAYS -f "%d-%m-%Y" "$START_DATE" +%d-%m-%Y)
    echo "Checking for date : $NEXT_DATE"
    output_file="result_${NEXT_DATE}_min-age_${MIN_AGE}.csv"
    echo "HOSPITAL_NAME, PINCODE, AVAILABLE_CAPACITY" > $output_file
    result_arr=""

    while read pincode; do
        result=$(curl --silent https://cdn-api.co-vin.in/api/v2/appointment/sessions/public/findByPin\?pincode\=$pincode\&date\=$NEXT_DATE | jq -r --arg MIN_AGE "$MIN_AGE" '.sessions[] | select(.min_age_limit == ($MIN_AGE | tonumber)) | select(.vaccine != "COVISHIELD") | [.name, .pincode, .available_capacity] | @csv')
        if [ ! -z "$result" ]; then
            result_arr="$result_arr\n$result"
        fi
    done < $pincode_file

    sorted_unique_result=$(echo -e $result_arr | sort | uniq)

    SAVEIFS=$IFS                                 # Save current IFS
    IFS=$'\n'                                    # Change IFS to new line
    sorted_unique_result=($sorted_unique_result) # split to array $names
    IFS=$SAVEIFS                                 # Restore IFS

   if (( ${#sorted_unique_result[@]} )); then
        echo "Vaccine slots available at ${#sorted_unique_result[@]} centres on $NEXT_DATE"
        printf '%s\n' "${sorted_unique_result[@]}"
   fi

done
