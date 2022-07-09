import numpy as np
import json
import numpy
import requests
import joblib
import math
import os
import pandas as pd
#import jaydebeapi
import os
import sys
#
redis_server_IP = "127.0.0.1"
redis_server_Port = 6379
#
if len(sys.argv) != 3 :
    print("Usage: python3 ./redis_loader_os_loz_six.py redis_server_IP  redis_server_Port.")
    quit()
else:
    redis_server_IP =  sys.argv[1]
    redis_server_Port = int(sys.argv[2])
#
print("Your redis_server_IP : ",redis_server_IP)
print("Your redis_server_Port : ",redis_server_Port)
#
# Mapper Preparing
def timeEncoder(X):
    X_hm = X['Time'].str.split(':', expand=True)
    d = pd.to_datetime(dict(year=X['Year'],month=X['Month'],day=X['Day'],hour=X_hm[0],minute=X_hm[1])).astype(int)
    return pd.DataFrame(d)

def amtEncoder(X):
    amt = X.apply(lambda x: x[1:]).astype(float).map(lambda amt: max(1,amt)).map(math.log)
    return pd.DataFrame(amt)

def decimalEncoder(X,length=5):
    dnew = pd.DataFrame()
    for i in range(length):
        dnew[i] = np.mod(X,10) 
        X = np.floor_divide(X,10)
    return dnew

def fraudEncoder(X):
    return np.where(X == 'Yes', 1, 0).astype(int)
#
from sklearn_pandas import DataFrameMapper
from sklearn.preprocessing import LabelEncoder
from sklearn.preprocessing import OneHotEncoder
from sklearn.preprocessing import FunctionTransformer
from sklearn.preprocessing import MinMaxScaler
from sklearn.preprocessing import LabelBinarizer
from sklearn.impute import SimpleImputer
#
mapper = joblib.load(open(os.path.join('./','fitted_mapper.pkl'),'rb'))
#
# Reading CSV
#
ddf = pd.read_csv('./test_220_100k_os.csv', dtype={"Merchant Name":"str"}, index_col='Index')
indices = np.loadtxt('test_220_100k.indices',dtype=int)
seq_length = 7
#
print(type(ddf),type(indices))#<class 'pandas.core.frame.DataFrame'> <class 'numpy.ndarray'> 
#
def gen_test_batch(ddf, mapper, indices):
    rows = indices.shape[0] 
    for i in range(rows - 1): 
        #print(type(indices[i]),indices[i])
        temp_input = ddf.loc[range(indices[i]-seq_length+1,indices[i]+1-1)]
        #print("temp_input",temp_input) 
        full_df = mapper.transform(temp_input)
        #print('full_df',full_df)
        tdf = full_df.drop(['Is Fraud?'],axis=1)
        #
        xbatch = tdf.to_numpy().reshape(1, seq_length-1, -1)
        xbatch_t = np.transpose(xbatch, axes=(1,0,2))
        data = json.dumps({"instances": xbatch_t.tolist()})
        #
        #
        yield indices[i], data
#
import redis
#Batch load into redis
r = redis.StrictRedis(host=redis_server_IP, port=redis_server_Port)
with r.pipeline(transaction=False) as p:
    i = 0 
    for index,data in gen_test_batch(ddf,mapper,indices):
        #print(index,len(data))
        #print(p.set(str(index),data))
        i = i + 1
        p.set(str(index),data)
        #
        if i % 1000 == 0 :
            print('Commited# ' + str(i))
            p.execute()
    #
    p.execute()
    #
quit()

