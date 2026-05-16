# MQTT TOPIC DESIGN


| 类型         | 方向        | Topic                                                         | Regex                                                                                                      | Event Type                | 发起方 | 必需性 |
|--------------|------------|---------------------------------------------------------------|------------------------------------------------------------------------------------------------------------|---------------------------|--------|--------|
| 属性上报     | 设备 → 云  | lf/v1/{pk}/{did}/property/up/post                             | `^lf/v1/(?P<product_key>[^/]+)/(?P<device_slug>[^/]+)/property/up/post$`                                   | `device.property.reported` | 设备   | 必需   |
| 属性上报 ack | 云 → 设备  | lf/v1/{pk}/{did}/property/down/post_reply                     | `^lf/v1/(?P<product_key>[^/]+)/(?P<device_slug>[^/]+)/property/down/post_reply$`                           | -                         | 云     | 可选¹  |
| 属性设置     | 云 → 设备  | lf/v1/{pk}/{did}/property/down/set                            | `^lf/v1/(?P<product_key>[^/]+)/(?P<device_slug>[^/]+)/property/down/set$`                                  | -                         | 云     | 必需   |
| 属性设置 ack | 设备 → 云  | lf/v1/{pk}/{did}/property/up/set_reply                        | `^lf/v1/(?P<product_key>[^/]+)/(?P<device_slug>[^/]+)/property/up/set_reply$`                              | `device.property.set.acknowledged` | 设备   | 必需   |
| 事件上报     | 设备 → 云  | lf/v1/{pk}/{did}/event/up/post                                | `^lf/v1/(?P<product_key>[^/]+)/(?P<device_slug>[^/]+)/event/up/post$`                                      | -                         | 设备   | 必需   |
| 事件上报 ack | 云 → 设备  | lf/v1/{pk}/{did}/event/down/post_reply                        | `^lf/v1/(?P<product_key>[^/]+)/(?P<device_slug>[^/]+)/event/down/post_reply$`                              | -                         | 云     | 可选¹  |
| 服务调用     | 云 → 设备  | lf/v1/{pk}/{did}/service/down/{service_name}                  | `^lf/v1/(?P<product_key>[^/]+)/(?P<device_slug>[^/]+)/service/down/(?P<service_name>[^/]+)$`               | -                         | 云     | 必需   |
| 服务调用 ack | 设备 → 云  | lf/v1/{pk}/{did}/service/up/{service_name}_reply              | `^lf/v1/(?P<product_key>[^/]+)/(?P<device_slug>[^/]+)/service/up/(?P<service_name>[^/]+)_reply$`           | -                         | 设备   | 必需   |
