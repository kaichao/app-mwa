#!/usr/bin/python3
#-*- coding:utf-8 -*-
import os
import sys
import json
import re
class messageRoute():
    def __init__(self):
        self.message = ""
        self.headers = ""
    
    def route_all(self,message,headers):
        headersstr = json.loads(headers)
        from_job=headersstr["from_job"]
        print(from_job)
        if from_job =="list-dir":
            messageRoute.filepack(message)
      
   
    @classmethod 
    def filepack(self,message):
        print('cccc'+message)
        try:
            print(message)
            data = json.loads(message)
            fmtFitsDataSet = '''
            {
                "datasetID":"zst:%s",
                "keyGroupRegex":"^.+/([0-9]+)_([0-9]+)_ch([0-9]+).dat.zst\$",
                "keyGroupIndex":"1,3,2",
                "sinkJob":"repack",
                "groupType":"V",
                "horizontalWidth":24,
                "verticalStart":%d,
                "verticalHeight":%d,
                "groupSize":%s,
                "interleaved":false
            }
            '''
            # 删除空格
            format = re.sub("\\s+", "", fmtFitsDataSet)
            s = format % (data["datasetID"], data["verticalStart"], data["verticalHeight"], os.getenv("NUM_PER_GROUP"))
            print(s)
            with open("/work/messages.txt", "a") as file:
                file.write("data-grouping-fits," + s)

        except ValueError:
            print(message)
            with open('/work/messages.txt', 'a') as file:
                arr1=message
                data='unpack,'+arr1
                file.write(data + '\n')
        
    
if __name__ == '__main__':
    parameter = sys.argv
    message=parameter[1]
    headers=parameter[2]
    print('message'+message)
    print('message'+message)
    #如何接收到headers
    w=messageRoute()
    w.route_all(message,headers)


