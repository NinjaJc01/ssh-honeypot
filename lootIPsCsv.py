import csv
from typing import List, Dict
import requests as r
import json
import time
from sys import stderr
from math import ceil
# https://ip-api.com/docs/api:batch


# class loginData:
#     def __init__(self, username, password, remoteIP, remoteVersion, timestamp, ipdata=None):
#         self.username = username
#         self.password = password
#         self.remoteIP = remoteIP
#         self.remoteVersion = remoteVersion
#         self.timestamp = timestamp
#         self.ipdata = ipdata

def divide_chunks(to_split, n) -> List:
    # looping till length l
    for i in range(0, len(to_split), n):
        yield to_split[i:i + n]


def getIpInfo(ips: List[str]) -> Dict:
    data = []
    ips = list(dict.fromkeys(ips))  # Remove duplicates
    print("Geolocating:\t",len(ips),"Addresses.\nEstimated time:",65 * ceil(len(ips) / 1400),"seconds")
    for ip in ips:
        data.append(  # Form Query for ip-api
            {"query": ip, "fields": "city,country,countryCode,region,regionName,city,isp,org,as,mobile,proxy,hosting,query"})
    data_chunked = divide_chunks(data,100)
    ip_data = list()
    backoff_counter = 1
    for datum in data_chunked:
        ##resp = r.post("http://ip-api.com/batch", json=datum)
        ##print("API Request number:\t",backoff_counter,file=stderr)
        ## We actually need to resend the request! Does this skip IPs?
        resp = ""
        if backoff_counter % 14 == 0: 
            print("Rate limit: Waiting 65 seconds",file=stderr)
            time.sleep(65)
            resp = r.post("http://ip-api.com/batch", json=datum)
        else:
            resp = r.post("http://ip-api.com/batch", json=datum)
        if resp.status_code == 429:
            print("Rate limit: Got 429, waiting",file=stderr)
            time.sleep(65)
            resp = r.post("http://ip-api.com/batch", json=datum)
        elif resp.status_code != 200:
            print(resp.status_code,resp.text,resp.headers)
        ip_data.extend(json.loads(str(resp.content, "utf-8")))
        backoff_counter += 1

    ip_dict = dict()
    for ip in ip_data:
        try:
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
        except KeyError as e:
            print(e,file=stderr)
            print(ip["query"],file=stderr)
    return ip_dict


def merge_ipdata_loot(loot: List[Dict], ipdata) -> List[Dict]:
    merged = list()
    for lootLine in loot:
        lootLine.update(ipdata[lootLine["remoteIP"]])
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
                "username":         record["username"].replace("\n",""),
                "password":         record["password"].replace("\n",""),
                "remoteIP":         record["remoteIP"].split(":")[0],
                "remoteVersion":    record["remoteVersion"].replace("\n",""),
                "timestamp":        record["timestamp"]
            }
            # if record["ip"].split(":")[0] == "51.91.157.101":
            #     ipList.append(data["remoteIP"])
            #     lootData.append(data)
            ipList.append(data["remoteIP"])
            lootData.append(data)
        return lootData, ipList

headers = ["username","password","remoteIP","remoteVersion","timestamp","country","countryCode","region","regionName","zip","isp","org","as","mobile","proxy","hosting"]
def main():
    lootData, ipList = loadCsv()
    ipInfo = getIpInfo(ipList)
    lootData = merge_ipdata_loot(lootData, ipInfo)
    myFile = open("loot2.csv","w",encoding='utf-8')
    writer = csv.writer(myFile, quoting=csv.QUOTE_ALL)
    writer.writerow(headers)
    print("Writing data to CSV")
    for data in lootData:
        writer.writerow(data.values())
    ##print(json.dumps(lootData))
    myFile.close()
main()
