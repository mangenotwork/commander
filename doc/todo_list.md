- 物理机可执行文件项目

  // TODO 路由分组
  // TODO master 与 slave 通信 设计为一个自定义协议， 并单独到包
  // 完整的数据包
  //1. 协议号
  //2. 消息头标识
  //3. 业务数据  一般就以下两种方式
  //文本方式：序列化文本文档text，或者json串，xml格式等；
  //二进制方式：常见的比如protobuf，thrift，kyro等；
  //4. 预留字段