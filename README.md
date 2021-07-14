# fabric-hub

## fabric同构间跨链网关

主要参考：
- [BitXHub白皮书](https://upload.hyperchain.cn/BitXHub%E7%99%BD%E7%9A%AE%E4%B9%A6.pdf)
- [合约跨链调用](https://wecross.readthedocs.io/zh_CN/latest/docs/dev/interchain.html)

网关间基于grpc通信（支持国密），数据传输过程中存在签名验签，保证数据的真实性。
