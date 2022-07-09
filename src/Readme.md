
# Readme
CCF sample start up in below steps:

##  0)  Prerequisites
### 0.1) Generate prerequisites by running Notebook from below link
[https://github.com/IBM/ai-on-z-fraud-detection/blob/main/ccf_220_keras_lstm_static-OS.ipynb](https://github.com/IBM/ai-on-z-fraud-detection/blob/main/ccf_220_keras_lstm_static-OS.ipynb)

Below files and directory will be generated:
         
- ****test_220_100k_os.csv**** : has all transctions to simulate 6 historyical transctions plus 1 new came in transction.    
It be used by REST API Container , when CICS only set the tx_index, REST API Container can find the body of the new came in transction.   
It be used by Reids Container to store the 6 historyical transctions with the index of the new came in transction as key.   
         
- ****test_220_100k.indices**** : has index number of the 1 new came in transction of test_220_100k_os.csv.    
It be used by REST API Container.    
It be used by Reids Conainter.       

- ****saved_models/P/ccf_220_keras_lstm_static/1**** : has the saved model to be used for inference. It be used by TFS Container.   
         
- ****fitted_mapper.pkl**** : Loading transctions from test_220_100k_os.csv into Redis, requried this file to mapping transctions into an TFS required JSON input.

### 0.2) Get required docker image or base image from IBM Registry

- ****TFS image**** : icr.io/ibmz/tensorflow-serving:2.7.0   
URL: https://ibm.github.io/ibm-z-oss-hub/containers/tensorflow-serving.html   
docker pull icr.io/ibmz/tensorflow-serving@sha256:8da2e8e497fc839a76cad33b16a76e1ed537730b762a4c7f17fb2673e27fcf55     
docker tag 27d0d64d5b2a icr.io/ibmz/tensorflow-serving:2.7.0    

- ****Redis image**** : icr.io/ibmz/redis:6.2.6
URL: https://ibm.github.io/ibm-z-oss-hub/containers/redis.html    
docker pull icr.io/ibmz/redis:6.2.6@sha256:ea17e0d3bff96aa84c458aee06404e1ea708eb5edc094bb47e38652ae7583f69   
- ****REST API Server base image**** : icr.io/ibmz/ubuntu:18.04    
URL: https://ibm.github.io/ibm-z-oss-hub/containers/ubuntu.html    
docker pull icr.io/ibmz/ubuntu:18.04@sha256:1185da02784dfbab9f3bee187311a2cb17efc4f8c027803a3c6b4a442a120e5c     


##  1)  Start up TFS docker container:

Start TFS docker image with saved model under directory saved_models/P/ccf_220_keras_lstm_static/1 
        
        docker run -t --rm -p 7501:8501 \
           -v "location_of_saved_models/P/ccf_220_keras_lstm_static/1_put_here:/models/ccf_220_os_z_lstm" \
           -e MODEL_NAME=ccf_220_os_z_lstm \
           --name ccf_220_os_z_tfs_v1 icr.io/ibmz/tensorflow-serving:2.4.0 &

##  2) Statr up Redis docker container:

### 2.1) Start up Redis container

        docker run  --rm -p 6579:6379 \
           -v "$(pwd)/redis_data_six/data:/data"  \
           --name redis626 -d icr.io/ibmz/redis:6.2.6 redis-server 

### 2.2) Load historical transctions from test_220_100k_os.csv into Redis

Make sure test_220_100k_os.csv , test_220_100k.indices and fitted_mapper.pkl under the same directory with redis_loader_os_loz_six.py

         python3 ./redis_loader_os_loz_six.py Your redis_server_IP Your redis_server_Port


##  3) Start up REST API server docker container:

### 3.1) Build REST API server docker image with Dockerfile
        
        docker build -t csl/api_svr_os_six:GA .    

### 3.2) Start REST API Server container 

         docker run  -p 8080:8080 -e REDISADD="Redis_IP:Redis_Port" -e TFSADD="TFS_IP:TFS_Port"  --name api_svr_six -d csl/api_svr_os_six:GA  
