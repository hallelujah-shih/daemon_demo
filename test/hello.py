import os
import sys
import time

if __name__ == "__main__":
    for i in xrange(100000000000000):
        if i % 2 == 0:
            print "ok: %s" % i
        else:
            sys.stderr.write("err: %s\n" % i)
        time.sleep(0.5)
