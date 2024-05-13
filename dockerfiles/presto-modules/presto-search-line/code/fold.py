#!/usr/bin/python3
import sys, os

def myexecute(cmd):
    print("'%s'"%cmd)
    os.system(cmd)

outsubs = False

if __name__ == '__main__':
    # read the file name to be processed
    filename = sys.argv[1]
    # get the file name without the extension
    basename = filename.split('.')[0]
    basename = basename.split('/')[-1]
    # read the candidate file
    candfile = sys.argv[2]
    with open(candfile, 'r') as f:
        lines = f.readlines()
        # parse the lines
        for line in lines:
            # if the line starts with format "  DM=a SNR=b" then skip it
            if line.startswith("  DM="):
                continue
            # else if the line starts with format "$basename_DM_xx_*:num" then split it into a list
            elif line.startswith(basename):
                params = line.split()
                dm = float(params[1])
                p = float(params[7])/1000.0

                myexecute("prepfold_gpu -noxwin -dm %.3f -p %.6f %s -o %s"%(dm, p, filename, basename))
