# Readme
CCF sample start up in below steps:


##  1)  Start up TFS docker container:

### 1.1) Get TFS base docker image

        docker pull icr.io/ibmz/tensorflow-serving:2.4.0@sha256:d232a0532342a29ed49d9cd61957793af07da6e8fba4d4c1da808124bb5909b7  

        docker tag 27d0d64d5b2a icr.io/ibmz/tensorflow-serving:2.4.0

### 1.2) Start TFS docker iamge with ccf_220_os_z_lstm model

        docker run -t --rm -p 7501:8501 \
           -v "$(pwd)/tfs_models/ccf_220_os_z_lstm:/models/ccf_220_os_z_lstm" \
           -e MODEL_NAME=ccf_220_os_z_lstm \
           --name ccf_220_os_z_tfs_v1 icr.io/ibmz/tensorflow-serving:2.4.0 &

##  2) Statr up Redis docker container:


### 2.1) Get image 

        https://ibm.github.io/ibm-z-oss-hub/containers/redis.html

        docker pull icr.io/ibmz/redis:6.2.6@sha256:ea17e0d3bff96aa84c458aee06404e1ea708eb5edc094bb47e38652ae7583f69

### 2.2) Start up image

        docker run  --rm -p 6579:6379 \
           -v "$(pwd)/redis_data_six/data:/data"  \
           --name redis626 -d icr.io/ibmz/redis:6.2.6 redis-server   

##  3) Statr up REST API server docker container:

### 3.1) Build REST API server docker image with Dockerfile
        
        docker build -t csl/api_svr_os_six:GA .    

### 3.2) Start up image 

         docker run  -p 8080:8080 -e REDISADD="9.30.43.79:6579" -e TFSADD="9.30.43.79:7501"  --name api_svr_six -d csl/api_svr_os_six:GA  
