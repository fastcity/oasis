# oasis 


#### 提供各个区块链链解析，上链等功能


#### Installation

1. [vct](https://github.com/abchain/fabric) 私有模式参见 [sdk](https://github.com/fastcity/oasis/tree/master/speed/config/dev/nodes)
2. eth 待开始
3. btc 待开始

#### Instructions

1. speed 提供区块解析服务，解析到的区块通过kafka 发送给 builder
2. builder 提供上链事务服务 同时接收speed的消息进行解析，更新用户的提交事务状态
3. api   对于builder 各个api的代理
4. space 对于各个api 接口的代理签名服务


+ 

#### api
##### api/v1/account 设置apikey secKey等 【PUT】
  - 请求参数
    + apiKey 用户 apiKey 请求接口签名用的 apiKey
    + secKey 用户 secKey 请求接口签名用的 secKey
    + cbUrl 用户 cbUrl 暂时无用


```
{
  "code": 0,
  "data": "success"
}
```
##### api/v1/account 设置apikey secKey等 【GET】
  - 请求参数
    + apiKey 用户 apiKey 请求接口签名用的 apiKey
  - 返回参数
    + 用户设置的签名参数

```
{
  "code": 0,
  "data": {
    "_id": "5d492028b11ef91b92248ca8",
    "apiKey": "gjf",
    "cbUrl": "gjf",
    "secKey": "gjf"
  }
}
```


#####  api/v1/createTransferTxData 添加签名事务 【POST】
  - 请求参数
    + from
    + to
    + value

  - response
    + hash 带签名hash
    + requestId 创建事务的数据库存储id
```
{
  "code": 0,
  "data": {
    "requestId": "5d68f77ffd54d4854a7ac833",
    "txData": {
      "hash": "C70EF7C6A794483DB2673FF8321F43B430F0EB8D724BB901E1EEA543DD0BE06F",
      "promise": {
        "Data": "91C8089F73EE10C23CE83A40D890BB83D7D5FCEFBF3F562C3DE12B46E0751091",
        "Nonce": "FBB77108AD1663CF76F94D9A31ECD69E374BA0FE",
        "txID": "pending"
      },
      "raw": "I::TOKEN.TRANSFER2:ChoKB0FCQ0hBSU4SD0F0b21pY0VuZXJneV92MRIGCJCLpOsFGhT7t3EIrRZjz3b5TZox7NaeN0ug/g==:CgEKEhYKFPmAE1K+T8V+y4ObshuwV2faNMOdGhYKFBJtq6Q46oTnxDwvVqMgDtZeNxs7"
    }
  }
}
```

##### api/v1/submitTx 提交事务 【POST】
  - 请求参数
    + signedRawTx 签名后的数据
    + requestId 创建事务的 requestId
    + sync  是否同步等待 true/fasle 

  - response
    + requestId 
    + sync 
    + txid sync 为true 会返回txid 链上txid
```
{
  "code": 0,
  "data": {
    "requestId": "5d68f77ffd54d4854a7ac833",
    "sync": true,
    "txid": "d25e246ce3f1c0831cad5bf4379e61974caeae09a8bc9a604d81d871677b2c06"
  },
  "msg": "create task success, the status of task will be notify"
}
```

##### api/v1/getTxStatus 获取上链状态 【POST】
   - 请求参数
     + requestId 
   - 返回参数
     + code 上链状态

|code|2|4|8|16|32|64|
|-|-|-|-|-|-|-|-|-
|status||api收到请求|发送到kafka(sync 为false时)|得到txid|已经上链|已经确认
```
{
  "code": 0,
  "data": {
    "__v": 0,
    "_account": "5ce276548aac08154973d9ce",
    "_id": "5ce75351d714cd34d47a00de",
    "chain": "ETH",
    "code": 0,
    "coin": "ETH",
    "createdAt": 1558664017402,
    "logs": [],
    "requestBody": {
      "apiKey": "12345678",
      "api_key": "12345678",
      "chain": "ETH",
      "coin": "ETH",
      "from": "0xfef1f3Dae07a41B00d785cF5af55C57F9145bca0",
      "signature": "62d9f8542402c72a7168c9442dfcdea9",
      "to": "0x0860123e5bc9bc6f40789e6f2929f7fdf35643ff",
      "value": "0.001"
    },
    "status": "ready",
    "updatedAt": 1558664017402
  }
}
```



##### /api/v1/balance 【GET】
  - 请求参数
    + address 签名后的数据
   

  - response
    + total 链上的余额
   
```
{
  "code": 0,
  "data": {
    "total": "100"
  }
}
```
