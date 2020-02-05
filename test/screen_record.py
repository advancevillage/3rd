#!/usr/bin/env python
# -*- coding: utf-8 -*-

"""
windows 无法使用外部signal 问题：
https://stackoverflow.com/questions/35772001/how-to-handle-the-signal-in-python-on-windows-machine
"""

import sys
import os
import time
import json
import getopt
import subprocess
import signal

process = None
pid = None
pid_file = './ffmpeg_pid.json'
# log.addLogFile("./test.log")

def get_ffmpeg_path():
    return os.path.join(os.path.dirname(os.path.abspath(__file__)), 'ffmpeg.exe')


def start(task_id, saved_path):
    # ffmpeg_path = 'C:\\ffmpeg-win64-static\\bin\\ffmpeg'
    ffmpeg_path = get_ffmpeg_path()
    # log.info("ffmpeg_path is: %s" % ffmpeg_path)

    arg = ' -f gdigrab -framerate 20 -i desktop -r 20 -y -vcodec mpeg4 "%s"' % saved_path
    cmd = ffmpeg_path + arg
    # devnull = open(os.devnull, 'w')
    # p = subprocess.Popen(cmd, stdin=subprocess.PIPE, stderr=devnull, shell=True)
    print(cmd)
    sys.exit()
    global process
    process = subprocess.Popen(cmd, stdin=subprocess.PIPE, stdout=subprocess.PIPE, shell=True)
    # write to file ffmpeg.pid
    if not os.path.exists(pid_file):
        with open(pid_file, 'w') as f:
            json.dump({task_id: process.pid}, f)
    else:
        with open(pid_file, 'r') as f:
            data = json.load(f)
            data[task_id] = process.pid
        with open(pid_file, 'w') as f:
            json.dump(data, f)
    process.wait()
    process.communicate(input='q'.encode())
    return process

#
# def handle_signal(signum, frame):
#     global process
#     print "receive:", signum
#     if signum == signal.SIGINT:
#
#         process.communicate(input='q')
#
#
# signal.signal(signal.SIGINT, handle_signal)
# print "pid", os.getpid()


def stop(pid):
    os.kill(pid, signal.SIGTERM)


def print_help():
    print ("usage: \n1.python screen_record.py start -t 4FDL098-GOLDEN-LOGIC -f D://record.mp4" \
          "\n2.python screen_record.py stop -p pid")


if __name__ == '__main__':
    try:
        opts, args = getopt.getopt(sys.argv[1:], '-h-t:-f:-p:', ['help', 'taskid=', 'filename=', 'pid='])
    except Exception:
        print_help()
        sys.exit()

    task_id = None
    file_name = None
    pid = None
    for k, v in opts:
        if k in ('-h', '--help'):
            print_help()
            sys.exit()
        if k in ('-t', '--taskid'):
            task_id = v
            print ("recording task_id is %s" % v)
        if k in ('-f', '--filename'):
            file_name = v
            print ("recording file is %s" % v)
        if k in ('-p', '--pid'):
            pid = int(v)
            print ("recording pid is %s" % v)
    for arg in args:
        if arg == 'start':
            if not task_id or not file_name:
                print_help()
                sys.exit()
            start(task_id, file_name)
            print ('recording finish...')
            with open(pid_file, 'r') as fp:
                data = json.load(fp)
                del data[task_id]
            with open(pid_file, 'w') as fp:
               json.dump(data, fp)
            break
        if arg == 'stop':
            if not pid:
                print_help()
                sys.exit()
            stop(pid)
            print ('recording stop.')
            break
    else:
        print_help()
        sys.exit()


