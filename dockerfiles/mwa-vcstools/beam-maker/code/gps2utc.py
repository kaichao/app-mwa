#!/usr/bin/python3

import numpy as np
from astropy.time import Time
import sys

gps = sys.argv[1]
t = Time(gps, format="gps")

print(t.fits)