import csv
from typing import List, Dict
import requests as r
import json
import time
from sys import stderr

def loadCsv() -> List[Dict]:
    f = open("cmd.csv", "r")
    with f:
        cmdData = list()
        reader = csv.DictReader(f)
        for record in reader:
            data = {
                "username":   record["user"].replace("\n",""),
                "remoteIP":   record["ip"].replace("\n",""),
                "command":    record["cmd"].replace("\n",""),
                "timestamp":  record["time"].replace("\n","")
            }
            # if record["ip"].split(":")[0] == "51.91.157.101":
            #     ipList.append(data["remoteIP"])
            cmdData.append(data)
        return cmdData

headers = ["user","ip","cmd","time"]
def main():
    lootData = loadCsv()
    myFile = open("cmd2.csv","w")
    writer = csv.writer(myFile, quoting=csv.QUOTE_ALL)
    writer.writerow(headers)
    print("Writing data to CSV")
    for data in lootData:
        writer.writerow(data.values())
    ##print(json.dumps(lootData))
    myFile.close()
main()
