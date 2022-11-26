#!/usr/bin/env bash
cd $(dirname $0)
go build 
for i in $(seq 0 10 100); do
	./icongenerator $i false | xml_pp > ../../refudeicons/scalable/status/refude_battery_discharging_${i}.svg
	./icongenerator $i true | xml_pp > ../../refudeicons/scalable/status/refude_battery_charging_${i}.svg
done
rm ./icongenerator

