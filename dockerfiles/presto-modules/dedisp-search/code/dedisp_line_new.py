#!/usr/bin/python3
import sys, os
###############################################
# This script is modified from disperse.py to #
# execute prepsubband_gpu on the given file   #
# using given settings                        #
###############################################

def myexecute(cmd):
    print("'%s'"%cmd)
    status = os.system(cmd)
    assert status == 0

outsubs = False

if __name__ == '__main__':
    # read the file name to be processed
    filename = sys.argv[1]
    basename = filename.split("/")[-1]
    # basename = basename.split(".")[0]
    f_settings_name = "/app/bin/MWA_DDplan.txt"
    # read arguments for accelsearch
    nsub = int(os.getenv("NSUB"))
    searchargs = os.getenv("SEARCHARGS")
    grpnum = int(os.getenv("GRPNUM"))
    ncpus = int(os.getenv("NCPUS"))
    workdir = os.getenv("WORK_DIR")
    target_mode=int(os.getenv("TARGET_MODE", 0))
    target_DM=float(os.getenv("TARGET_DM", 0))

    print(searchargs)
    # read rfi file name from command line
    if len(sys.argv) > 2:
        f_rfi_name = sys.argv[2]
        use_rfi = True
    else:
        use_rfi = False
    # read lines from settings file
    with open(f_settings_name, "r") as f_settings:
        lines = f_settings.readlines()
        
    # the settings file should have format
#  Low DM    High DM     dDM  DownSamp  dsubDM   #DMs    DMs/call  calls  WorkFract
#    1.000    43.420    0.01       2     14.14    4242    1414       3     0.6114
#   43.420    67.980    0.01       4     24.56    2456    2456       1     0.177
#   67.980   117.100    0.02       8     49.12    2456    2456       1     0.08849
#  117.100   343.300    0.05      16    113.10    4524    2262       2     0.0815
    dDMs = []
    dsubDMs = []
    startDMs = []
    downsamps = []
    subcalls = []
    dmspercalls = []
    lines.pop(0)
    
    for line in lines:
        line = line.split()
        dDMs.append(float(line[2]))
        dsubDMs.append(float(line[4]))
        downsamps.append(int(line[3]))
        subcalls.append(int(line[7]))
        startDMs.append(float(line[0]))
        dmspercalls.append(int(line[6]))

# Loop over the DDplan plans
    cnt = 0
    linenum = 0
    try:
        for dDM, dsubDM, dmspercall, downsamp, subcall, startDM in zip(dDMs, dsubDMs, dmspercalls, downsamps, subcalls, startDMs):
            # Loop over the number of calls
            linenum += 1
            for ii in range(subcall):
                cnt += 1
                if target_mode == 0 and cnt != grpnum:
                    continue
                subDM = startDM + (ii+0.5)*dsubDM
                loDM = startDM + ii*dsubDM
                if target_mode != 0:
                    hiDM = startDM + (ii+1)*dsubDM
                    if loDM > target_DM or hiDM < target_DM:
                        continue
                    else:
                        subDM = target_DM
                        loDM = target_DM - 5*dDM
                        dmspercall = 10
                if outsubs:
                    # Get our downsampling right
                    subdownsamp = downsamp // 2
                    datdownsamp = 2
                    if downsamp < 2: subdownsamp = datdownsamp = 1
                    # First create the subbands
                    myexecute("prepsubband_gpu -cuda %d -ncpus %d -sub -subdm %.2f -noclip -nsub %d -downsamp %d -o %s %s/*.fits" %
                            (0, ncpus, subDM, nsub, subdownsamp, basename, filename))
                    # And now create the time series
                    subnames = basename+"_DM%.2f.sub[0-9]*"%subDM
                    myexecute("prepsubband_gpu -cuda %d -ncpus %d -lodm %.2f -dmstep %.2f -noclip -numdms %d -downsamp %d -o %s %s/*.fits" %
                            (0, ncpus, loDM, dDM, dmspercall, datdownsamp, basename, subnames))
                elif use_rfi:
                    myexecute("prepsubband_gpu -cuda %d -ncpus %d -nsub %d -lodm %.2f -dmstep %.2f -noclip -zerodm -numdms %d -downsamp %d -mask %s -o %s %s/*.fits" %
                            (0, ncpus, nsub, loDM, dDM, dmspercall, downsamp, f_rfi_name, basename, filename))
                
                else:
                    myexecute("prepsubband_gpu -cuda %d -ncpus %d -nsub %d -lodm %.2f -dmstep %.2f -noclip -zerodm -numdms %d -downsamp %d -o %s %s/*.fits" %
                            (0, ncpus, nsub, loDM, dDM, dmspercall, downsamp, basename, filename))
                myexecute("echo %d > linenum.txt" %(linenum))

                
                # call prepsubband_gpu to process the file
        # myexecute("date --iso-8601=ns >> /work/timestamps.txt")
        # myexecute("realfft *.dat")
        # myexecute("rm -f *.dat")
        # myexecute("ls *.fft | xargs -n 1 accelsearch_gpu_4 -cuda 0 " + searchargs)
        # myexecute("rm -f *.fft")
        # myexecute("date --iso-8601=ns >> %s/timestamps.txt", workdir)
    except:
        exit(1)

