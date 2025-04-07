#!/usr/bin/python3
import sys, os
import concurrent.futures

def myexecute(cmd):
    print("'%s'"%cmd)
    status = os.system(cmd)

def process_line(line):
    if line.startswith("  DM="):
        return None
    elif line.startswith(basename):
        params = line.split()
        dm = params[1]
        p = float(params[7]) / 1000.0
        candname = params[0].split(":")
        accelfile = candname[0]
        candnum = int(candname[1])
        if p > 0.015:
            nbins = 100
            ntimechunk = 120
            dmstep = 1
        else:
            nbins = 50
            ntimechunk = 40
            dmstep = 3
        return p, nbins, ntimechunk, dmstep, dm, accelfile, candnum

def execute_commands(data):
    p, nbins, ntimechunk, dmstep, dm, accelfile, candnum = data
    myexecute("prepfold -noxwin  -accelfile %s.cand -accelcand %d \
                 -pstep 1 -pdstep 2 -npfact 2 -ndmfact 3 \
                -n %d -npart %d -dmstep %d -o %s_DM%s_%.6fs %s_DM%s.dat " % (accelfile, candnum,
                 nbins, ntimechunk, dmstep, basename, dm, p, basename, dm))

if __name__ == '__main__':
    filename = sys.argv[1]
    basename = filename.split('/')[-2]
    candfile = sys.argv[2]

    with open(candfile, 'r') as f:
        lines = f.readlines()

    # Parse lines in the main thread
    parsed_data = []
    for line in lines:
        data = process_line(line)
        if data:
            parsed_data.append(data)

    # Define the maximum number of threads for command execution
    max_threads = 4

    # Create a ThreadPoolExecutor with a maximum of max_threads
    with concurrent.futures.ThreadPoolExecutor(max_workers=max_threads) as executor:
        # Submit each task to the executor
        futures = [executor.submit(execute_commands, data) for data in parsed_data]

        # Wait for all tasks to complete
        for future in concurrent.futures.as_completed(futures):
            # Get the result of each task (if needed)
            result = future.result()

    print(len(parsed_data))
