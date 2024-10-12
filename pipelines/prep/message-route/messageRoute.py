#!/usr/bin/python3
#-*- coding:utf-8 -*-
import os
import sys
import json
import re
import subprocess
class messageRoute():
    def __init__(self):
        self.message = ""
        self.headers = ""

    def route_all(self,message,headers):

        if headers =="null" or "from_job" not in headers:
            print("222222222222222")
            messageRoute.fileready(message)
        else:
            try:
                headersstr = json.loads(headers)
                from_job=headersstr["from_job"]
                from_ip=headersstr["from_ip"]
                print(from_job)
                if from_job == "dir-list":
                    messageRoute.unpack(message)
                if from_job == "unpack":
                    messageRoute.isredist(message,from_ip)

                if from_job == "fits-redist":
                    messageRoute.repack(message,from_ip)
                if from_job == "repack":
                    messageRoute.rclonecopy(message)
            except json.JSONDecodeError as e:
                print("Invalid JSON format in headers:", e)

    @classmethod
    def unpack(self,message):
        #执行解包操作
        command = f"scalebox task add -sink-job=unpack {message}"
        result=subprocess.run(command, shell=True)
        if result.returncode == 0:
            print("命令执行成功")
            return result.returncode
        else:
            print(f"命令执行失败，返回码为: {result.returncode}")
            return result.returncode
    @classmethod
    def rclonecopy(self,message):
        #执行解包操作
        command = f"scalebox task add -sink-job=rclone-copy {message}"
        result=subprocess.run(command, shell=True)
        if result.returncode == 0:
            print("命令执行成功")
            return result.returncode
        else:
            print(f"命令执行失败，返回码为: {result.returncode}")
            return result.returncode   
    @classmethod
    def isredist(self,message,from_ip):

        # 判断是否包含 "ch"
        if "ch" in message:
            # 使用正则表达式提取数字
            match = re.search(r'ch(\d+)', message)
            if match:
                number = int(match.group(1))
                ip_ranges = {
                    "10.255.11.1": (109, 110),
                    "10.255.11.2": (111, 112),
                    "10.255.11.3": (113, 114),
                    "10.255.11.4": (115, 116),
                    "10.255.11.5": (117, 118),
                    "10.255.11.6": (119, 120),
                    "10.255.11.7": (121, 122),
                    "10.255.11.8": (123, 124),
                    "10.255.11.9": (125, 126),
                    "10.255.11.10": (127, 128),
                    "10.255.11.11": (129, 130),
                    "10.255.11.12": (131, 132)
                }
                if from_ip in ip_ranges:
                    low, high = ip_ranges[from_ip]
                    if low <= number <= high:
                        messageRoute.repack(message, from_ip)
                        print(f"Processed for {from_ip}: {low} to {high}")
                    else:
                        for ip, (low_range, high_range) in ip_ranges.items():
                            if low_range <= number <= high_range:
                                SOURCE_URL = f"root@{from_ip}/data/mwa"
                                to_ip = ip
                                break
                        command = "scalebox"
                        arguments = ["task", "add", "--sink-job", "fits-redist", "--header", f"source_url={SOURCE_URL}", "--to-ip", to_ip,  message]
                        result = subprocess.run([command] + arguments, capture_output=True, text=True)
                        if result.returncode == 0:
                            print("命令执行成功")
                        else:
                            print(f"命令执行失败，返回码为: {result.returncode}")
                            return result.returncode
        else:
            print("文件名不包含 'ch'")
            return 0
            
    @classmethod
    def repack(self,message,from_ip):
       
        print(f"repack11111111111111111:{message}{from_ip}")
       # matches = re.findall(r'(\d+)_(\d+)+', message)
        matches = re.findall(r'(\d+)_([\d]+)_(ch\d+)', message)
        datasetid=matches[0][0]
        print(datasetid)
        given_number=int(matches[0][1])
        print(given_number)
        ch_number=matches[0][2]
        print(ch_number)
        end_st=int(os.environ.get("end_st"))
        star_st=int(os.environ.get("star_s"))
        size=int(os.environ.get("size"))
        if given_number < star_st:
            print("小于初始值")
        elif given_number > end_st:
            print("大于结束值")
        else:
            group_index = (given_number - star_st) // size
            group_start = star_st + group_index * size
            group_end = group_start + size - 1
            if group_end > end_st:
                group_end = end_st
            sema = f'prep_ready:{datasetid}/{group_start}_{group_end}/{ch_number}'
            if sema:
                print(f"{message} 在组 {sema} 中")
                command = f"scalebox semaphore countdown {sema}"
                #output = subprocess.check_output(command, shell=True)
                result=subprocess.run(command, shell=True, stdout=subprocess.PIPE)
                if result.returncode == 0:
                    n = int(result.stdout.decode())
                    print(f"获取的值 n 为: {n}")
                    if(n==0):
                        print("开始打包")
                        sendmessage = f"{datasetid}/{group_start}_{group_end}/{ch_number}"
                        print(sendmessage)
                        command = "scalebox"
                        arguments = ["task", "add", "--sink-job", "repack", "--to-ip", from_ip, sendmessage]
                        result = subprocess.run([command] + arguments, capture_output=True, text=True)
                        print(result.returncode)
                        return result.returncode
                else:
                    print("命令执行失败")
                    print("错误信息:", result.stderr)
                    return result.returncode
            else:
                print(f"{message} 不在任何组中")

    @classmethod
    def fileready(self,message):
        print(message)

        t=int(os.environ.get("size"))
        print(t)
        dataset_id=os.environ.get("dataset_id")
        end_st=int(os.environ.get("end_st"))
        star_st=int(os.environ.get("star_s"))
        #star_st=1253471954
        # fir_c=109#起始通道号
        # sum_c=24#通道总数
        sum_st=end_st-star_st+1
        #商式div
        div=sum_st//t
        print(div)
        #余氏mod
        mod=sum_st%t
        print(mod)
        my_dict = {}
        for j in range(109,133):
            for i in range(div):
                semaphore="prep_ready:"+str(dataset_id)+'/'+str(star_st+i*t)+'_'+str(star_st+(i+1)*t-1)+"/ch"+str(j)
                # print(semaphore)
                command = f"scalebox semaphore create {semaphore} {t}"
                subprocess.run(command, shell=True)
            if mod > 0: 
                i=div
                semaphore="prep_ready:"+str(dataset_id)+'/'+str(star_st+i*t)+'_'+str(end_st)+"/ch"+str(j)
                command = f"scalebox semaphore create {semaphore} {mod}"
                subprocess.run(command, shell=True, stdout=subprocess.PIPE, stderr=subprocess.PIPE)
        print(os.environ.get("message"))
        with open('/work/messages.txt', 'a') as file:
            arr=os.environ.get("message")
            data='dir-list,'+arr
            file.write(data + '\n')

        return 0
if __name__ == '__main__':
    parameter = sys.argv
    message=parameter[1]
    headers=parameter[2]
    print('message'+message)
    print('headers'+headers)
    #如何接收到headers
    w=messageRoute()
    w.route_all(message,headers)















