#!/bin/bash

ZIP_FILE="vaccination-centres.zip"
OUTPUT_DIR="./results"

rm $ZIP_FILE 2> /dev/null
rm -rf $OUTPUT_DIR 2> /dev/null
mkdir -p $OUTPUT_DIR

# Lucknow, Ranchi, Kolkata
districts=( 670 240 725 )

for district in ${districts[@]}
do
   echo "Checking for district id $district"
   csvfile=$(./vaccine-status byDistrict $district)
   if [ -f "$csvfile" ]; then
      echo "$csvfile"
      mv "$csvfile" "$OUTPUT_DIR/$district.csv"
   fi
done

if [ $( ls $OUTPUT_DIR/ | wc -l ) -eq 0 ]; then
    echo "No available slot found"
    exit 0
fi

# trigger email
echo "Triggering email ...."

zip $ZIP_FILE $OUTPUT_DIR/*
base64_encoded_data=$( base64 $ZIP_FILE )

printf -v data '{"personalizations": [{"bcc": [{"email": "siddharth98391@gmail.com"}, {"email": "ironmanlucknow@gmail.com"}, {"email": "alokdutt@safearth.in"},{"email": "akshatsrivastava.lko@gmail.com"}]}],"from": {"email": "lko-vaccination-alert@vaccine-baba.com"},"subject":"Vaccination slots available in Lucknow","content": [{"type": "text/html","value": "Hey buddy,<br><br>Please find attached the details of vaccine availability at different centres in Lucknow.<br> Stay safe, stay vaccinated.<br><br>Cheers,<br>Vaccine Baba"}], "attachments": [{"content": "%s", "type": "text/plain", "filename": "%s"}]}' "$base64_encoded_data" "$ZIP_FILE"

echo "$data" > request.json

curl --request POST \
  --url https://api.sendgrid.com/v3/mail/send \
  --header "authorization: Bearer $SENDGRID_API_KEY" \
  --header 'Content-Type: application/json' \
  --data @request.json

rm request.json

echo "Sent email"
