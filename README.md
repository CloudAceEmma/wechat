## wechat
### sendmsg: POST
#### Request
##### HTTP request
```
POST http://SERVER_HOST_IP:8080/sendmsg
```
##### Content type
application/json
##### Parameters
| Parameter name  | Value  | Description              |
| --------------- | -------| ------------------------ |
| TextMsg         | string | message content          |
| NickName        | string | messaging target nickname|
##### Response
```
{"message": "message content"}
```
If successful, this method returns message with content "success",
else, it returns the error message.

### sendfile: POST
#### Request
##### HTTP request
```
POST http://SERVER_HOST_IP:8080/sendfile
```
##### Content type
multipart/form-data
##### Form Parameters
| Parameter key   | Value  | Description                  |
| --------------- | -------| ---------------------------- |
| file            | bytes  | file content and name to send|
| NickName        | string | messaging target nickname    |
##### Response
```
{"message": "message content"}
```
If successful, this method returns message with content "success",
else, it returns the error message.

### Examples
```
curl http://127.0.0.1:8080/sendfile -H 'Content-Type: multipart/form-data' -F 'NickName=test001' -F 'file=@media/gopher.jpg'
```
```
curl http://127.0.0.1:8080/sendmsg -H 'Content-Type: application/json' -d '{"TextMsg": "hello from wechat bot", "NickName": "test001"}'
```
