FROM icr.io/ibmz/ubuntu:18.04
RUN  rm -rf /home/ccf_os
RUN  mkdir  /home/ccf_os
COPY test_220_100k.csv  /home/ccf_os/
COPY mapper.so   /home/ccf_os/
COPY api_svr_os_six_ga  /home/ccf_os/
CMD  cd /home/ccf_os/;./api_svr_os_six_ga
