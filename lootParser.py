import csv
from typing import List, Dict
import requests as r
import json
import time
from sys import stderr
# https://ip-api.com/docs/api:batch


def divide_chunks(to_split, n) -> List:
    # looping till length l
    for i in range(0, len(to_split), n):
        yield to_split[i:i + n]


def getIpInfo(ips: List[str]) -> Dict:
    data = []
    ips = list(dict.fromkeys(ips))  # Remove duplicates
    for ip in ips:
        data.append(  # Form Query for ip-api
            {"query": ip, "fields": "city,country,countryCode,region,regionName,city,isp,org,as,mobile,proxy,hosting,query"})
    data_chunked = divide_chunks(data,15)
    ip_data = list()
    backoff_counter = 0
    for datum in data_chunked:
        resp = r.post("http://ip-api.com/batch", json=datum)
        print("API Request number:\t",backoff_counter,file=stderr)
        if backoff_counter % 15 == 0 and backoff_counter != 0: 
            print("Backing off for a minute due to rate limiting",file=stderr)
            time.sleep(60)
        ip_data.extend(json.loads(str(resp.content, "utf-8")))
        backoff_counter += 1
    ip_dict = dict()

    for ip in ip_data:
        ip_dict[ip["query"]] = {
            "country": ip["country"],
            "countryCode": ip["countryCode"],
            "region": ip["region"],
            "regionName": ip["regionName"],
            "zip": ip["zip"],
            "isp": ip["isp"],
            "org": ip["org"],
            "as": ip["as"],
            "mobile": ip["mobile"],
            "proxy": ip["proxy"],
            "hosting": ip["hosting"],
        }
    return ip_dict


def merge_ipdata_loot(loot: List[Dict], ipdata) -> List[Dict]:
    merged = list()
    for lootLine in loot:
        lootLine["ipdata"] = ipdata[lootLine["remoteIP"]]
        merged.append(lootLine)
    return merged


def loadCsv() -> (List[Dict], List[str]):
    f = open("loot.csv", "r")
    with f:
        lootData = list()
        ipList = []
        reader = csv.DictReader(f)
        for record in reader:
            data = {
                "username":         record["user"],
                "password":         record["pass"],
                "remoteIP":         record["ip"].split(":")[0],
                "remoteVersion":    record["ver"],
                "timestamp":        record["time"]
            }
            # if record["ip"].split(":")[0] == "51.91.157.101":
            #     ipList.append(data["remoteIP"])
            #     lootData.append(data)
            ipList.append(data["remoteIP"])
            lootData.append(data)
        return lootData, ipList


def main():
    lootData, ipList = loadCsv()
    ipInfo = getIpInfo(ipList)
    lootData = merge_ipdata_loot(lootData, ipInfo)
    print(json.dumps(lootData))


main()
