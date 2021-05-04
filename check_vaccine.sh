#!/bin/bash

pincode_file='lucknow-pincode.txt'
START_DATE=04-05-2021
# MIN_AGE=18
MIN_AGE=45
ZIP_FILE="vaccination-centres.zip"
OUTPUT_DIR="./results"

rm $ZIP_FILE 2> /dev/null
rm -rf $OUTPUT_DIR 2> /dev/null
mkdir -p $OUTPUT_DIR

for i in {0..0}
do
    if [[ "$OSTYPE" == "darwin"* ]]; then
        DAYS="${i}d"
        NEXT_DATE=$(date -j -v +$DAYS -f "%d-%m-%Y" "$START_DATE" +%d-%m-%Y)
    else
        NEXT_DATE=$(date +%d-%m-%Y -d "$DATE + $i day")
    fi

    echo "Checking for date : $NEXT_DATE"
    output_file="$OUTPUT_DIR/result_${NEXT_DATE}_min-age_${MIN_AGE}.csv"
    header="HOSPITAL_NAME, PINCODE, AVAILABLE_CAPACITY, VACCINE_NAME"

    echo "" > $output_file

    while read pincode; do
        result=$(curl --silent https://cdn-api.co-vin.in/api/v2/appointment/sessions/public/findByPin\?pincode\=$pincode\&date\=$NEXT_DATE | jq -r --arg MIN_AGE "$MIN_AGE" '.sessions[] | select(.min_age_limit == ($MIN_AGE | tonumber)) | [.name, .pincode, .available_capacity, .vaccine] | @csv')
        if [ ! -z "$result" ]; then
            echo "$result" >> $output_file
        fi
    done < $pincode_file

    sorted_unique_result=$(cat $output_file | sort -u | uniq)

   if [ ! -z "$sorted_unique_result" ]; then
        sed "1s/^/$header\n/" $output_file
        body=$( cat $output_file )

        # printing on console        
        echo $subject
        echo "$body"
        echo ""
   else
        rm $output_file
   fi

done

# trigger email
echo "Triggering email ...."

zip $ZIP_FILE $OUTPUT_DIR/*
base64_encoded_data=$( base64 $ZIP_FILE )

printf -v data '{"personalizations": [{"to": [{"email": "siddharth98391@gmail.com"}]}],"from": {"email": "lko-vaccination-alert@vaccine-baba.com"},"subject":"Vaccination slots available in Lucknow","content": [{"type": "text/html","value": "Hey buddy,<br><br>Please find attached the details of vaccine availability at different centres in Lucknow.<br> Stay safe, stay vaccinated.<br><br>Cheers,<br>Vaccine Baba"}], "attachments": [{"content": "%s", "type": "text/plain", "filename": "%s"}]}' "$base64_encoded_data" "$ZIP_FILE"

curl --request POST \
  --url https://api.sendgrid.com/v3/mail/send \
  --header "authorization: Bearer $SENDGRID_API_KEY" \
  --header 'Content-Type: application/json' \
  --data "$data"

echo "Sent email"
