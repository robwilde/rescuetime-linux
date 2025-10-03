Testing activation number 6 in `./rescuetime-auth.http`


Request:

```json
POST https://api.rescuetime.com/activate
Content-Type: application/x-www-form-urlencoded
Content-Length: 51
User-Agent: IntelliJ HTTP Client/GoLand 2025.2.2
Accept-Encoding: br, deflate, gzip, x-gzip
Accept: */*
Cookie: ahoy_visitor=fb24729d-f11b-4369-933b-ae14b962d695

username=robert%40mrwilde.com&password=U44bcV8lGb5q

###
```

Response:

```
---
c:
- 0
- RT:ok
account_key: 186c3aa4fddc9204ea5e6cb2dfb50fa2
key: 186c3aa4fddc9204ea5e6cb2dfb50fa2
```