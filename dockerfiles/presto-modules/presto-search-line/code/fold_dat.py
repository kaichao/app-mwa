#!/usr/bin/python3
import sys, os

def myexecute(cmd):
    print("'%s'"%cmd)
    os.system(cmd)

outsubs = False

if __name__ == '__main__':
    # read the file name to be processed
    cnt = 0
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
                dm = params[1]
                p = float(params[7])/1000.0
                cnt += 1
                if p > 0.0015:
                    nbins=100
                    ntimechunk=120
                    dmstep=1
                else:
                    nbins=50
                    ntimechunk=40
                    dmstep=3
                myexecute("prepfold -noxwin -p %.6f -pstep 1 -pdstep 2 -npfact 2 -ndmfact 3 \
                            -n %d -npart %d -dmstep %d -o %s_DM%s %s_DM%s.dat "%(p, 
                             nbins, ntimechunk, dmstep, basename, dm, basename, dm))
    print(cnt)
